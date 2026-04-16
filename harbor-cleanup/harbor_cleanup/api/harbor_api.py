"""
Harbor Cleanup Tool - Harbor API Module

This module handles all API communications with Harbor registry.
"""

import time
import os
import json
import re
import requests
from requests.auth import HTTPBasicAuth
import urllib3
from concurrent.futures import ThreadPoolExecutor, as_completed
from urllib.parse import quote

from ..core.config import (
    HARBOR_URL, USERNAME, PASSWORD, VERIFY_SSL, 
    API_RETRY_COUNT, API_RETRY_DELAY, 
    MAX_PAGE_WORKERS, PARALLEL_PAGE_THRESHOLD, ENABLE_PARALLEL,
    VERBOSE_ARTIFACTS, DRY_RUN, GC_DELETE_UNTAGGED, GC_WORKERS, logger
)

urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)

# Global session configuration without connection pooling to avoid pool exhaustion
# Connection pooling is disabled due to high concurrency causing pool exhaustion
_session = requests.Session()
_session.verify = False  # SSL verification handled by VERIFY_SSL config

# Log the session configuration
logger.info("🔗 HTTP session configured without connection pooling to prevent pool exhaustion")
logger.info("🔒 CSRF token handling enabled for write operations")

# Global cache for repository list to avoid repeated API calls
_repositories_cache = None

# Global cache for CSRF token
_csrf_token = None

def clear_csrf_token():
    """Clear cached CSRF token to force refresh on next request"""
    global _csrf_token
    _csrf_token = None
    logger.debug("🔒 Cleared cached CSRF token")

def get_csrf_token(force_refresh=False):
    """Get CSRF token from Harbor for write operations"""
    global _csrf_token
    
    # Return cached token if available and not forcing refresh
    if _csrf_token and not force_refresh:
        return _csrf_token
    
    try:
        # Get CSRF token by making a GET request to Harbor's web interface
        csrf_url = f"{HARBOR_URL}/c/login"
        logger.debug(f"🔒 Getting CSRF token from {csrf_url}")
        
        response = _session.get(
            csrf_url,
            auth=HTTPBasicAuth(USERNAME, PASSWORD),
            verify=VERIFY_SSL,
            timeout=30
        )
        
        # Extract CSRF token from response headers or cookies
        csrf_token = None
        
        # Try to get from X-Harbor-CSRF-Token header
        if 'X-Harbor-CSRF-Token' in response.headers:
            csrf_token = response.headers['X-Harbor-CSRF-Token']
            logger.debug("🔒 Found CSRF token in X-Harbor-CSRF-Token header")
        
        # Try to get from cookies
        elif '_csrf' in response.cookies:
            csrf_token = response.cookies['_csrf']
            logger.debug("🔒 Found CSRF token in _csrf cookie")
        
        # Try to extract from HTML content (last resort)
        else:
            import re
            csrf_match = re.search(r'name="_csrf"[^>]*value="([^"]*)"', response.text)
            if csrf_match:
                csrf_token = csrf_match.group(1)
                logger.debug("🔒 Found CSRF token in HTML content")
        
        if csrf_token:
            _csrf_token = csrf_token
            logger.debug(f"🔒 Successfully obtained CSRF token: {csrf_token[:8]}...")
            return csrf_token
        else:
            logger.warning("⚠️  Could not find CSRF token in response")
            return None
            
    except Exception as e:
        logger.warning(f"⚠️  Failed to get CSRF token: {e}")
        return None

def get_session_info():
    """Get information about the current session configuration"""
    session_info = {
        'connection_pooling': 'disabled',
        'ssl_verification': VERIFY_SSL,
        'session_configured': True,
        'csrf_token_cached': _csrf_token is not None
    }
    logger.debug(f"🔗 Session info: {session_info}")
    return session_info

def encode_repo_name(repo_name):
    """
    Encode repository name for Harbor API URLs.
    According to Harbor API docs, if repository name contains slash, 
    encode it twice over with URL encoding. e.g. a/b -> a%2Fb -> a%252Fb
    """
    if '/' in repo_name:
        # Double encode for repository names with slashes
        return quote(quote(repo_name, safe=''), safe='')
    else:
        # Single encode for simple names (though not strictly necessary)
        return quote(repo_name, safe='')

def api_request_with_retry(method, url, **kwargs):
    """
    Wrapper for requests with retry logic and exponential backoff
    
    Args:
        method: HTTP method ('GET', 'POST', 'DELETE', etc.)
        url: The URL to request
        **kwargs: Additional arguments to pass to requests
    
    Returns:
        requests.Response object or None if all retries failed
    """
    method = method.upper()

    # Build request executor using session for all methods with proper CSRF handling
    def execute_request():
        # Set up kwargs with proper SSL verification
        safe_kwargs = dict(kwargs)
        safe_kwargs['verify'] = VERIFY_SSL
        
        # Get headers
        headers = dict(safe_kwargs.get("headers") or {})
        
        if method in ("POST", "PUT", "PATCH", "DELETE"):
            # For write operations, get and include CSRF token
            csrf_token = get_csrf_token()
            if csrf_token:
                headers["X-Harbor-CSRF-Token"] = csrf_token
                logger.debug(f"🔒 Added CSRF token to {method} request")
            else:
                logger.warning(f"⚠️  No CSRF token available for {method} request - this may fail")
        
        safe_kwargs["headers"] = headers
        return _session.request(method, url, **safe_kwargs)
    
    for attempt in range(API_RETRY_COUNT + 1):  # +1 because attempt 0 is the first try
        try:
            if attempt > 0:
                # Calculate exponential backoff delay
                delay = API_RETRY_DELAY * (2 ** (attempt - 1))
                logger.debug(f"🔄 Retry attempt {attempt}/{API_RETRY_COUNT} for {method} {url} after {delay:.1f}s delay...")
                time.sleep(delay)
            
            response = execute_request()
            
            # Check for CSRF token errors and retry with fresh token
            if (response.status_code == 403 and 
                method in ("POST", "PUT", "PATCH", "DELETE") and
                "CSRF token not found" in response.text):
                
                logger.warning(f"🔒 CSRF token error detected, refreshing token...")
                clear_csrf_token()  # Clear cached token
                
                # Retry with fresh token if we haven't exceeded retry count
                if attempt < API_RETRY_COUNT:
                    continue
            
            # Success cases - return immediately  
            if response.status_code < 500:  # Don't retry client errors (4xx), only server errors (5xx)
                return response
            
            # Server error - log and potentially retry
            if attempt < API_RETRY_COUNT:
                logger.warning(f"⚠️  {method} {url} failed with {response.status_code} (attempt {attempt + 1}/{API_RETRY_COUNT + 1}), retrying...")
                logger.warning(f"📄 Response: {response.text[:300]}{'...' if len(response.text) > 300 else ''}")
            else:
                logger.error(f"❌ {method} {url} failed with {response.status_code} after {API_RETRY_COUNT + 1} attempts")
                logger.error(f"📄 Final response: {response.text}")
                return response  # Return the failed response so caller can handle it
                
        except requests.exceptions.RequestException as e:
            if attempt < API_RETRY_COUNT:
                logger.warning(f"⚠️  {method} {url} failed with exception: {e} (attempt {attempt + 1}/{API_RETRY_COUNT + 1}), retrying...")
            else:
                logger.error(f"❌ {method} {url} failed with exception after {API_RETRY_COUNT + 1} attempts: {e}")
                raise  # Re-raise the exception after all retries are exhausted
    
    return None  # Should never reach here

def get_cached_repositories(project, silent=False):
    """Get repositories with caching to avoid repeated API calls"""
    global _repositories_cache
    
    if _repositories_cache is None:
        if not silent:
            logger.info("🔄 Fetching repository list for first time...")
        _repositories_cache = list_repositories(project)
    else:
        if not silent:
            logger.info(f"✅ Using cached repository list ({len(_repositories_cache)} repositories)")
    
    return _repositories_cache

def list_repositories(project):
    """Fetch all repositories in a project"""
    page = 1
    page_size = 100
    repositories = []
    logger.info(f"🔍 Fetching repositories from project {project}...")
    
    while True:
        logger.debug(f"Fetching repositories page {page}...")
        r = api_request_with_retry(
            "GET",
            f"{HARBOR_URL}/api/v2.0/projects/{project}/repositories",
            params={"page": page, "page_size": page_size},
            auth=HTTPBasicAuth(USERNAME, PASSWORD),
            verify=VERIFY_SSL
        )
        if r.status_code != 200:
            logger.error(f"❌ Error fetching repositories: {r.status_code} - {r.text}")
            break
        
        batch = r.json()
        if not batch:
            if page == 1:
                logger.warning("No repositories found")
            else:
                logger.debug("No more repositories found")
            break
        
        repositories.extend(batch)
        logger.debug(f"Found {len(batch)} repositories on page {page}...")
        logger.info(f"Total repositories: {len(repositories)}")
        page += 1
    
    logger.info(f"✅ Found {len(repositories)} repositories.")
    return repositories

def get_repository_info(project, repo_name):
    """Get repository information including artifact count"""
    # Strip project prefix if present, preserving nested paths (e.g. "razorpay/cache/infra-tools" -> "cache/infra-tools")
    prefix = f"{project}/"
    repo = repo_name[len(prefix):] if repo_name.startswith(prefix) else repo_name
    
    logger.info(f"🔍 Getting repository info for {project}/{repo}...")
    
    # Use the repository details API first (more efficient)
    encoded_repo = encode_repo_name(repo)
    repo_details_url = f"{HARBOR_URL}/api/v2.0/projects/{project}/repositories/{encoded_repo}"
    logger.debug(f"🔗 Repository details URL: {repo_details_url}")
    
    r = api_request_with_retry(
        "GET",
        repo_details_url,
        auth=HTTPBasicAuth(USERNAME, PASSWORD),
        verify=VERIFY_SSL
    )
    
    if r.status_code == 401:
        logger.warning("⚠️  Repository details API unauthorized, falling back to repository list API...")
        logger.debug(f"   Repository details API failed for {project}/{repo}, using list fallback")
        # Fallback to cached repository list API
        repositories = get_cached_repositories(project)
        
        logger.debug(f"🔧 Looking for exact match: {project}/{repo}")
        
        # Look for repository - prefer exact project/repo match, fallback to short name match
        exact_match = None
        fallback_match = None
        
        for repo_data in repositories:
            repo_full_name = repo_data['name']
            # Extract repo name from full name (e.g., "razorpay/authz" -> "authz")
            repo_short = repo_full_name.split('/')[-1] if '/' in repo_full_name else repo_full_name
            
            # Check for exact repository name match first
            if repo_full_name == repo:
                # Direct match (e.g., looking for "authz" and found "authz")
                exact_match = repo_data
                logger.debug(f"🔧 Found exact match: {repo_full_name} (direct name match)")
                break  # Exact match takes priority
            elif repo_full_name == f"{project}/{repo}":
                # Project/repo match (e.g., looking for "authz" and found "razorpay/authz")
                if exact_match is None:  # Only if no direct match found yet
                    exact_match = repo_data
                    logger.debug(f"🔧 Found exact match: {repo_full_name} (project/repo match)")
                    break
            elif repo_short == repo and fallback_match is None:
                # Store first fallback but continue looking for exact match
                fallback_match = repo_data
                logger.debug(f"🔧 Found fallback match: {repo_full_name} (matching '{repo}')")
        
        # Use exact match if found, otherwise use fallback
        result_match = exact_match or fallback_match
        
        # Return match if found
        if result_match:
            artifact_count = result_match.get('artifact_count', 0)
            match_type = "exact match" if exact_match else "fallback match"
            logger.info(f"✅ Found repository {result_match['name']} with {artifact_count} artifacts (via cached list - {match_type})")
            return {
                'artifact_count': artifact_count,
                'name': result_match['name'],
                'id': result_match.get('id'),
                'project_id': result_match.get('project_id'),
                'creation_time': result_match.get('creation_time'),
                'update_time': result_match.get('update_time'),
                'pull_count': result_match.get('pull_count', 0)
            }
        
        logger.error(f"❌ Repository {repo_name} not found in project {project}")
        return None
    elif r.status_code == 404:
        logger.warning(f"⚠️  Repository {project}/{repo} not found via details API, trying list API fallback...")
        # Fallback to cached repository list API for 404s too
        repositories = get_cached_repositories(project)
        for repo_data in repositories:
            repo_full_name = repo_data['name']
            repo_short = repo_full_name.split('/')[-1] if '/' in repo_full_name else repo_full_name
            if repo_short == repo or repo_full_name == repo:
                artifact_count = repo_data.get('artifact_count', 0)
                logger.info(f"✅ Found {repo_full_name} with {artifact_count} artifacts (via list API fallback)")
                return repo_data
        logger.error(f"❌ Repository {repo_name} not found in project {project}")
        return None
    elif r.status_code != 200:
        logger.error(f"❌ Error getting repository info: {r.status_code} - {r.text}")
        logger.warning("⚠️  Falling back to repository list API...")
        # Try fallback for other errors too
        repositories = get_cached_repositories(project)
        for repo_data in repositories:
            repo_full_name = repo_data['name']
            repo_short = repo_full_name.split('/')[-1] if '/' in repo_full_name else repo_full_name
            if repo_short == repo or repo_full_name == repo:
                artifact_count = repo_data.get('artifact_count', 0)
                logger.info(f"✅ Found {repo_full_name} with {artifact_count} artifacts (via list API fallback)")
                return repo_data
        logger.error(f"❌ Repository {repo_name} not found after fallback")
        return None
    
    repo_info = r.json()
    artifact_count = repo_info.get('artifact_count', 0)
    logger.info(f"✅ Repository {project}/{repo} has {artifact_count} artifacts (via repository details API)")
    logger.debug(f"📊 Repository info: {repo_info}")
    
    return repo_info

def list_artifacts(project, repo, artifact_count=None):
    """Fetch all artifacts from a repository with optional parallel page processing"""
    page_size = 100
    artifacts = []
    
    # Determine if we should use parallel processing based on artifact count
    # Option 1: Threshold-based (current smart approach)
    use_parallel_pages_threshold = (
        ENABLE_PARALLEL and 
        artifact_count is not None and 
        artifact_count > PARALLEL_PAGE_THRESHOLD
    )
    
    # Option 2: Always parallel (if you want to force it)
    use_parallel_pages_always = (
        ENABLE_PARALLEL and 
        artifact_count is not None and 
        artifact_count > 0  # Always use parallel if we have any artifacts
    )
    
    # Choose your strategy (uncomment the one you want)
    # use_parallel_pages = use_parallel_pages_threshold  # Current smart approach
    use_parallel_pages = use_parallel_pages_always   # Always parallel approach
    
    logger.info(f"🔧 Parallel page fetching decision for {project}/{repo}:")
    logger.info(f"   - ENABLE_PARALLEL: {ENABLE_PARALLEL}")
    logger.info(f"   - artifact_count: {artifact_count}")
    logger.info(f"   - PARALLEL_PAGE_THRESHOLD: {PARALLEL_PAGE_THRESHOLD}")
    logger.info(f"   - use_parallel_pages: {use_parallel_pages}")
    
    if use_parallel_pages:
        estimated_pages = (artifact_count + page_size - 1) // page_size  # Ceiling division
        logger.info(f"🔍 Fetching {artifact_count} artifacts from {project}/{repo} using parallel processing...")
        logger.info(f"🚀 Estimated {estimated_pages} pages, using parallel page fetching...")
        
        def fetch_page(page_num):
            """Fetch a single page of artifacts"""
            try:
                encoded_repo = encode_repo_name(repo)
                r = api_request_with_retry(
                    "GET",
                    f"{HARBOR_URL}/api/v2.0/projects/{project}/repositories/{encoded_repo}/artifacts",
                    params={"page": page_num, "page_size": page_size, "with_label": True},
                    auth=HTTPBasicAuth(USERNAME, PASSWORD),
                    verify=VERIFY_SSL
                )
                if r.status_code == 200:
                    batch = r.json()
                    return {
                        'page': page_num,
                        'artifacts': batch,
                        'success': True,
                        'count': len(batch) if batch else 0
                    }
                else:
                    return {
                        'page': page_num,
                        'artifacts': [],
                        'success': False,
                        'error': f"{r.status_code} - {r.text}",
                        'count': 0
                    }
            except Exception as e:
                return {
                    'page': page_num,
                    'artifacts': [],
                    'success': False,
                    'error': str(e),
                    'count': 0
                }
        
        # Use parallel processing to fetch all pages
        max_pages_to_try = estimated_pages + 2  # Add buffer for safety
        logger.info(f"⚡ Starting parallel page fetch with {min(MAX_PAGE_WORKERS, max_pages_to_try)} workers for {max_pages_to_try} pages...")
        
        with ThreadPoolExecutor(max_workers=min(MAX_PAGE_WORKERS, max_pages_to_try)) as executor:
            # Submit tasks for all estimated pages
            page_futures = []
            for page_num in range(1, max_pages_to_try + 1):
                future = executor.submit(fetch_page, page_num)
                page_futures.append(future)
            
            # Collect results as they complete
            page_results = []
            completed_pages = 0
            for future in as_completed(page_futures):
                try:
                    result = future.result()
                    page_results.append(result)
                    completed_pages += 1
                    if completed_pages % 10 == 0 or completed_pages <= 10:  # Log first 10 and every 10th
                        logger.info(f"📄 Parallel fetch progress for {project}/{repo}: {completed_pages}/{max_pages_to_try} pages completed")
                except Exception as exc:
                    logger.error(f"❌ Exception fetching page: {exc}")
            
            # Sort results by page number and add successful ones
            page_results.sort(key=lambda x: x['page'])
            
            for result in page_results:
                if result['success'] and result['artifacts']:
                    artifacts.extend(result['artifacts'])
                    logger.debug(f"✅ Fetched page {result['page']}: {result['count']} artifacts")
                elif result['success'] and not result['artifacts']:
                    # Empty page means we've reached the end
                    logger.debug(f"🏁 Reached end at page {result['page']}")
                    break
                else:
                    logger.error(f"❌ Failed to fetch page {result['page']}: {result.get('error', 'Unknown error')}")
                    # Continue with other pages even if one fails
        
        logger.info(f"🏁 Parallel fetching complete. Total artifacts: {len(artifacts)}")
        
    else:
        # Use sequential processing for smaller repositories or when parallel is disabled
        if artifact_count is not None:
            logger.info(f"🔍 Fetching {artifact_count} artifacts from {project}/{repo} using sequential processing...")
        else:
            logger.info(f"🔍 Fetching artifacts from {project}/{repo} (count unknown, using sequential)...")
        
        page = 1
        encoded_repo = encode_repo_name(repo)
        while True:
            logger.debug(f"Fetching page {page}...")
            r = api_request_with_retry(
                "GET",
                f"{HARBOR_URL}/api/v2.0/projects/{project}/repositories/{encoded_repo}/artifacts",
                params={"page": page, "page_size": page_size, "with_label": True},
                auth=HTTPBasicAuth(USERNAME, PASSWORD),
                verify=VERIFY_SSL
            )
            if r.status_code != 200:
                logger.error(f"❌ Error: {r.status_code} - {r.text}")
                break
                
            batch = r.json()
            if not batch:
                if page == 1:
                    logger.warning("No artifacts found")
                else:
                    logger.debug("No more artifacts found")
                break
                
            artifacts.extend(batch)
            logger.debug(f"Found {len(batch)} artifacts on page {page}...")
            logger.info(f"Total artifacts in {project}/{repo}: {len(artifacts)}")
                
            if len(batch) < page_size:  # Last page
                break
                    
            page += 1
    
    logger.info(f"✅ Found {len(artifacts)} artifacts in {project}/{repo}.")
    return artifacts

def check_artifact_exists(project, repo, reference):
    """Check if an artifact exists in Harbor"""
    try:
        encoded_repo = encode_repo_name(repo)
        r = api_request_with_retry(
            "GET",
            f"{HARBOR_URL}/api/v2.0/projects/{project}/repositories/{encoded_repo}/artifacts/{reference}",
            auth=HTTPBasicAuth(USERNAME, PASSWORD),
            verify=VERIFY_SSL
        )
        return r.status_code == 200
    except Exception:
        return False

def delete_artifact(project, repo, reference):
    """Delete an artifact from Harbor using digest (tag deletion is deprecated)"""
    # Determine if reference is a tag or digest
    is_digest = reference.startswith('sha256:')
    ref_type = "digest" if is_digest else "tag"
    ref_display = reference[:12] + "..." if is_digest else reference
    
    if DRY_RUN:
        logger.debug(f"🗑️  [DRY RUN] Would delete artifact {ref_display} ({ref_type})")
        return True
    
    # Check if artifact exists before attempting deletion
    if not check_artifact_exists(project, repo, reference):
        logger.warning(f"⚠️  Artifact {ref_display} not found before deletion attempt")
        return True  # Consider this success since artifact is gone
    
    logger.info(f"🗑️  Deleting artifact {ref_display} ({ref_type})")
    logger.debug(f"📋 Delete parameters: project={project}, repo={repo}, reference={reference}")
    
    try:
        encoded_repo = encode_repo_name(repo)
        delete_url = f"{HARBOR_URL}/api/v2.0/projects/{project}/repositories/{encoded_repo}/artifacts/{reference}"
        logger.debug(f"🔗 Delete URL: {delete_url}")
        
        r = api_request_with_retry(
            "DELETE",
            delete_url,
            auth=HTTPBasicAuth(USERNAME, PASSWORD),
            verify=VERIFY_SSL
        )
        
        if r.status_code in [200, 202]:
            logger.info(f"✅ Deleted {ref_display}")
            return True
        elif r.status_code == 404:
            logger.warning(f"⚠️  Artifact {ref_display} not found (already deleted?)")
            return True  # Consider this a success since the artifact is gone
        elif r.status_code == 412:
            logger.error(f"❌ Cannot delete {ref_display}: Artifact is referenced by other manifests or tags")
            logger.error(f"   💡 This artifact may be part of a multi-arch image or referenced by multiple tags")
            return False
        elif r.status_code == 403:
            logger.error(f"❌ Permission denied deleting {ref_display}: Check Harbor user permissions")
            logger.error(f"Error message - {r.text}")
            return False
        elif r.status_code == 500:
            logger.error(f"❌ Harbor server error deleting {ref_display}: {r.text}")
            logger.error(f"   💡 Possible causes:")
            logger.error(f"   - Harbor's custom tag deletion logic failure (regCli.DeleteTag + tagCtl.Delete)")
            logger.error(f"   - Multi-arch manifest internal reference issues")
            logger.error(f"   - Harbor database consistency issues")
            logger.error(f"   - Storage backend problems") 
            logger.error(f"   - Concurrent access conflicts")
            if not is_digest:
                logger.error(f"   🔧 Harbor team should review custom tag deletion implementation")
            return False
        else:
            logger.error(f"❌ Failed to delete {ref_display}: HTTP {r.status_code} - {r.text}")
            return False
    except Exception as e:
        logger.error(f"❌ Exception deleting {ref_display}: {e}")
        return False



def delete_artifact_with_fallback(project, repo, digest, tags, artifact_data=None):
    """
    Delete an artifact using digest-based deletion only.
    
    Note: Tag-based deletion has been removed due to unreliable behavior with Harbor's 
    custom tag deletion logic. Only digest-based deletion is now used for consistency
    and reliability.
    
    Args:
        project: Harbor project name
        repo: Repository name
        digest: Artifact digest (sha256:...)
        tags: Tags associated with artifact (not used for deletion, kept for compatibility)
        artifact_data: Optional artifact metadata for logging
    
    Returns:
        bool: True if deletion successful, False otherwise
    """
    # Log artifact type for debugging
    if artifact_data and artifact_data.get('references'):
        references = artifact_data.get('references', [])
        logger.info(f"🔗 Multi-arch manifest with {len(references)} child references (Harbor will auto-cleanup children)")
    else:
        logger.debug(f"📋 Simple manifest")
    
    # Log associated tags for reference (but don't use them for deletion)
    if tags:
        tag_list = tags if isinstance(tags, list) else [tags]
        logger.debug(f"📋 Associated tags: {', '.join(tag_list)}")
    
    # Get artifact size for logging
    artifact_size = 0
    if artifact_data:
        artifact_size = artifact_data.get("size", 0)
    
    # Import format_size locally to avoid circular imports
    from ..utils.formatting import format_size
    size_info = f" ({format_size(artifact_size)})" if artifact_size > 0 else ""
    
    # Delete using digest only
    logger.info(f"🗑️  Deleting by digest {digest[:12]}{size_info}...")
    success = delete_artifact(project, repo, digest)
    
    if success:
        logger.info(f"✅ Successfully deleted artifact {digest[:12]} by digest{size_info}!")
        if artifact_data and artifact_data.get('references'):
            logger.info(f"   🔗 Harbor automatically cleaned up {len(artifact_data.get('references', []))} child manifests")
        return True
    else:
        logger.error(f"❌ Failed to delete artifact {digest[:12]} by digest{size_info}")
        if artifact_data and artifact_data.get('references'):
            logger.error(f"   💡 Multi-arch manifest deletion failed - may require manual intervention")
        else:
            logger.error(f"   💡 Simple manifest deletion failed - may be Harbor permission, state, or logic issue")
        return False

def trigger_gc():
    """Trigger Harbor Garbage Collection to reclaim disk space.
    Sends a manual GC schedule with fixed parameters.
    """
    if DRY_RUN:
        logger.info("🚮 [DRY RUN] Would trigger Harbor Garbage Collection...")
        return True
    
    logger.info("🚮 Triggering Harbor Garbage Collection...")

    # Base payload with manual schedule and configurable parameters
    payload = {
        "schedule": {"type": "Manual"},
        "parameters": {
            "delete_untagged": GC_DELETE_UNTAGGED,
            "dry_run": False,
            "time_window": 0,
            "workers": GC_WORKERS
        }
    }
    
    try:
        r = api_request_with_retry(
            "POST",
            f"{HARBOR_URL}/api/v2.0/system/gc/schedule",
            json=payload,
            auth=HTTPBasicAuth(USERNAME, PASSWORD),
            verify=VERIFY_SSL,
            timeout=30  # Add timeout for safety
        )
        
        # Harbor GC API typically returns 201 (Created) for successful scheduling
        if r.status_code in [200, 201]:
            logger.info("✅ GC triggered successfully.")
            return True
        elif r.status_code == 409:
            logger.warning("⚠️  GC is already running or scheduled.")
            return True  # Not really an error
        else:
            logger.error(f"❌ Failed to trigger GC: {r.status_code} - {r.text}")
            return False
            
    except requests.exceptions.Timeout:
        logger.error("❌ Failed to trigger GC: Request timeout")
        return False
    except requests.exceptions.RequestException as e:
        logger.error(f"❌ Failed to trigger GC: Network error - {e}")
        return False
    except Exception as e:
        logger.error(f"❌ Failed to trigger GC: Unexpected error - {e}")
        return False 