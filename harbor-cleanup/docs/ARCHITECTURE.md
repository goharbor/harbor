# Harbor Cleanup Tool - Modular Architecture

This document explains the modular structure of the Harbor Cleanup Tool, organized into focused modules for better maintainability and readability.

## Architecture Overview

The tool has been broken down into the following modules:

```
tools/harbor-cleanup/
├── harbor_cleanup/              # Main package
│   ├── __init__.py             # Package initialization  
│   ├── core/                   # Core business logic
│   │   ├── __init__.py
│   │   ├── config.py           # Configuration and environment variables
│   │   ├── processor.py        # Repository processing and parallel execution logic
│   │   └── retention_policy.py # Retention policy logic and categorization
│   ├── api/                    # External interfaces
│   │   ├── __init__.py
│   │   └── harbor_api.py       # Harbor API communication layer
│   └── utils/                  # Common utilities
│       ├── __init__.py
│       ├── formatting.py       # Common utility functions and helpers
│       ├── reporting.py        # Output formatting and summary generation
│       └── slack_notifier.py   # Slack notifications for cleanup summaries
├── main.py                     # Main entry point with modular architecture
├── tests/                      # Test modules
│   ├── __init__.py
│   ├── test_modules.py         # Module import test script
│   └── test_structure.py       # Package structure validation
├── docs/                       # Documentation
│   ├── README.md               # Main documentation
│   └── ARCHITECTURE.md         # This file
├── docker/                     # Container files
│   ├── Dockerfile
│   └── .dockerignore
├── scripts/                    # Helper scripts
│   └── setup_local_env.sh
└── requirements.txt            # Dependencies
```

## Package Responsibilities

### 📦 `harbor_cleanup/` - Main Package
The root package that exports all functionality and provides clean imports.

### 🔧 `harbor_cleanup/core/` - Core Business Logic Package

#### `config.py`
- **Purpose**: Centralized configuration management
- **Contains**: Environment variable handling, logging setup, credential validation
- **Key exports**: `logger`, `HARBOR_URL`, `PROJECT`, `DRY_RUN`, etc.

#### `retention_policy.py`
- **Purpose**: Artifact categorization and retention logic
- **Contains**: `RetentionPolicy` class with age/count policies and label matching
- **Key features**: Configurable retention rules, artifact categorization

#### `processor.py`
- **Purpose**: Repository processing logic and parallel execution
- **Contains**: Main processing functions, parallel/sequential execution modes
- **Key features**: Multi-threaded processing, progress tracking, artifact analysis

### 🌐 `harbor_cleanup/api/` - External Interfaces Package

#### `harbor_api.py`
- **Purpose**: All Harbor API communications
- **Contains**: API requests with retry logic, repository/artifact fetching, deletion
- **Key features**: Parallel page fetching, caching, error handling

### 🛠️ `harbor_cleanup/utils/` - Common Utilities Package

#### `formatting.py`
- **Purpose**: Common utility functions and helpers
- **Contains**: Signal handlers, formatting functions, shutdown management
- **Key exports**: `format_size()`, `format_duration()`, `setup_signal_handlers()`

#### `reporting.py`
- **Purpose**: Output formatting and summary generation
- **Contains**: Startup banner, progress reports, final summaries
- **Key features**: Rich formatting, timing statistics, cleanup summaries

#### `slack_notifier.py`
- **Purpose**: Slack integration for notifications
- **Contains**: SlackNotifier class, message formatting, API communication
- **Key features**: Start/end notifications, error alerts, rich message blocks

### 🚀 `main.py` - Entry Point
- **Purpose**: Clean main entry point that orchestrates everything
- **Contains**: High-level flow control, error handling, graceful shutdown
- **Key features**: Simple, readable main function

### 🧪 `tests/` - Test Package
- **Purpose**: Validation and testing of the package structure
- **Contains**: Module import tests, structure validation tests
- **Key features**: Automated validation of package organization

## Benefits of Refactoring

### ✅ **Improved Maintainability**
- Each package has a single, clear responsibility
- Logical grouping of related functionality
- Easier to find and modify specific features
- Reduced risk of unintended side effects

### ✅ **Better Testability**
- Individual packages can be tested in isolation
- Mocking dependencies is much easier
- Test coverage can be more granular
- Dedicated tests/ directory for all validation

### ✅ **Enhanced Readability**
- Smaller, focused files are easier to understand
- Clear package hierarchy and separation of concerns
- Professional Python package structure
- Self-documenting organization

### ✅ **Easier Development**
- Multiple developers can work on different packages
- Reduced merge conflicts through separation
- Clearer code review scope
- Standard Python import patterns

### ✅ **Reusability**
- Individual packages can be imported and used elsewhere
- API layer can be reused for other Harbor tools
- Utilities can be shared across projects
- Clean package imports (`from harbor_cleanup.api import harbor_api`)

### ✅ **Professional Structure**
- Follows Python packaging best practices
- Proper `__init__.py` files for clean imports
- Organized documentation in docs/ directory
- Separated Docker files and scripts

## Usage

### Running the Tool
```bash
# Use the main entry point
python main.py

# Run as a Python package
python -c "from harbor_cleanup import main; main()"

# Run tests to validate structure
PYTHONPATH=. python tests/test_structure.py
```

### Testing the Module Structure
```bash
# Test that all modules import correctly
python test_modules.py
```

### Environment Variables
All the same environment variables work as before:
```bash
export HARBOR_USERNAME="your_username"
export HARBOR_PASSWORD="your_password"
export HARBOR_PROJECT="razorpay"
export DRY_RUN="true"
export REPO_LIST="relay,ledger,authz"
# ... etc
```

## Migration Guide

### For Users
- **No changes required** - all environment variables and functionality remain the same
- Use `main.py` as the single entry point
- All CLI behavior is identical

### For Developers
- **Import specific modules** instead of copying code
- **Extend functionality** by modifying individual modules
- **Add new features** by creating new modules or extending existing ones

## Example: Adding a New Feature

To add a new retention policy type:

1. **Modify `retention_policy.py`** - Add new policy logic
2. **Update `config.py`** - Add new environment variables
3. **Extend `processor.py`** - Apply new policy in processing
4. **Update `reporting.py`** - Include new policy in summaries

This modular approach makes adding features much more straightforward!

## Backward Compatibility

- **Original `main.py` is preserved** for compatibility
- **All existing functionality** works exactly the same
- **All environment variables** are supported
- **All CLI behaviors** are identical

## Future Improvements

With this modular structure, future improvements become much easier:

- **Plugin system** for custom retention policies
- **Multiple Harbor instances** support
- **Database storage** for cleanup history
- **Web UI** for monitoring and configuration
- **Prometheus metrics** export
- **Custom notification** systems

The modular architecture provides a solid foundation for these enhancements! 