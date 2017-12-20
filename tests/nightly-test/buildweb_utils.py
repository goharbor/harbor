import time, urllib2, json
import xml.etree.ElementTree as ET
import re

maxBuild = 1
class BuildWebUtil:
    """Interact with build web."""

    def __init__(self):
        self.delimiter = '#' * 80
        self.lineSep = '\r\n'

    @staticmethod
    def get_resource_list(url, times=1):
        """Get build info data"""
        try:
            build_web = 'http://buildapi.eng.vmware.com'
            url = '%s%s' % (build_web, url)
            print 'Fetching %s ...\r\n' % url
            ret = urllib2.urlopen(url)
            status = int(ret.code)
            if status != 200:
                print('HTTP status %d', status)
                raise Exception('Error: %s' % data['http_response_code']['message'])
            content = ret.read()
            if json:
                data = json.loads(content)
            else:
                data_dict = {}
                data_list = []
                for i in content.replace(' ', '').split('[{')[1].split('}]')[0].split('},{'):
                    for j in i.split(','):
                        data_dict[j.split(':')[0].strip().strip('"')] = j.split(':')[1].strip().strip('"')
                    data_list.append(data_dict)
                    data_dict = {}
                data_dict['_list'] = data_list
                data = data_dict
        except Exception, e:
            print(e)
            print(url)
            times += 1
            time.sleep(5)
            if times < 10:
                BuildWebUtil.get_resource_list(url, times)
        return data

    @staticmethod
    def get_latest_recommend_build(product, branch, build_type='ob'):

        print("Max Search Build Number: [%d]" % maxBuild)
        ret = ''
        found_recommend = False
        url = '/%s/build/?' \
              'product=%s&' \
              'branch=%s&' \
              '_limit=%d&' \
              '_order_by=-id&' \
              'buildstate__in=succeeded,storing&' \
              'buildtype__in=release,beta' \
              % (build_type, product, branch, maxBuild)
        data = BuildWebUtil.get_resource_list(url)

        for item in data['_list']:
            build_id = item['id']
            """Save the latest build if there's no "recommended" build"""
            if ret == '': ret = str(build_id)
            qa_rrl = '/%s/qatestresult/?build=%s' % (build_type, build_id)
            qa_data = BuildWebUtil.get_resource_list(qa_rrl)
            for result in qa_data['_list']:
                if result['qaresult'] == "recommended":
                    ret = str(build_id)
                    found_recommend = True
            if found_recommend:
                break
        return ret

    @staticmethod
    def get_latest_build(product, branch, build_type='ob'):
        ''' only get the latest release build '''
        url = '/%s/build/?' \
              'product=%s&' \
              'branch=%s&' \
              '_limit=%d&' \
              '_order_by=-id&' \
              'buildstate__in=succeeded,storing&' \
              'buildtype__in=release' \
              % (build_type, product, branch, 1)
        """print 'url is %s' % url"""
        data = BuildWebUtil.get_resource_list(url)
        return data['_list'][0]['id']

    @staticmethod
    def get_deliverable_list_by_build_id(build_id, build_type='ob'):
        url = '/%s/deliverable/?build=%s' % (build_type, build_id)
        data = BuildWebUtil.get_resource_list(url)
        if data is not None:
            return data['_list']
        return None

    @staticmethod
    def get_deliverable_by_build_id(build_id, target_pattern,
                                    build_type='ob'):
        print('Entering {0}'.format(
            BuildWebUtil.get_deliverable_by_build_id.__name__))
        buildInfoUrl = "http://buildweb.eng.vmware.com/{0}/api/legacy" \
                       "/build_info/?build={1}".format(build_type, build_id)

        ret = urllib2.urlopen(buildInfoUrl)
        status = int(ret.code)
        if status != 200:
            logging.error('HTTP status %d', status)
            raise Exception('Error: %s' % ret.msg)
        infoRoot = ET.fromstring(ret.read())

        for c1 in infoRoot[0]:
            matched = re.search(target_pattern, c1.attrib['url'])
            if matched:
                print(c1.attrib['url'])
                return c1.attrib['url']

    @staticmethod
    def get_version_by_build_num(build_id, build_type='ob'):
        url = '/%s/build/%s' % (build_type, build_id)
        data = BuildWebUtil.get_resource_list(url)
        if data is not None:
            return data['version']
        return None

    def download_by_build_num(self, build_id, build_type, product, system_arch, retry=True):
        """
        :param build_id:
        :param product:
        :param system_arch: x86 or x64 -- this is a predefined agent vm arch, as this script will not communicate with
                                          the vm machine. Thus, just hard code before run.

                                          Actually, the script already enhanced to communicate with VM machine with
                                          vsphere API. As no bug so far, will not update the code here.
        :param retry:

       :rtype : str, build local path. Should be in the temp folder.
       """
        print("Entering download_by_build_num")
        print(build_id)
        print(build_type)
        print(product)
        print(system_arch)

        builds = BuildWebUtil.get_deliverable_list_by_build_id(build_id, build_type)
        # for build in builds:
           # print(builds)

        target_path = None
        target_pattern = get_pattern_product(product, system_arch)
        print(target_pattern)

        assert (builds != '')
        assert (target_pattern != '')

        for build in builds:
            # print(build)
            matched = re.search(target_pattern, build['path'])
            if matched:
                target_path = build['_download_url']
                break
        if not target_path:
            target_path = self.get_deliverable_by_build_id(build_id,
                                                    target_pattern, build_type)
            # target_path = 'http://build-squid.eng.vmware.com/build/mts' \
            #               '/release' \
            #              '/bora-{0}/publish/VMware-Horizon-Client-4.6.0' \
            #              '-{1}.exe'.format(build_id, build_id)
        print(target_path)

        """download build to temp folder."""
        try:
            local_file_path = os.path.join(tempfile.gettempdir(), os.path.basename(target_path))
            # print('Downloading... : ' + str(build_id))
            if not os.path.exists(local_file_path):
                urllib.urlretrieve(target_path, local_file_path)
            print('The Build Was Downloaded Successfully on the Local : [%s]' % local_file_path)
        except Exception, e:
            print(e)
            if not retry:
                raise DownloadError("Build ID: [%s] Download Error." % str(build_id))
            """will retry one more to handle any exception"""
            print('Get Download Exception, Will Try One More Time and Stop the Workflow If Meet It Twice.')
            self.download_by_build_num(build_id, product, system_arch, False)

        return local_file_path


# buildwebutil = BuildWebUtil()
# build_id=buildwebutil.get_latest_recommend_build('harbor_build', 'master')
# print "Got url:"+buildwebutil.get_deliverable_by_build_id(build_id, '.*.ovf')

