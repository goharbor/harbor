"""
Command-line interface to the Harbor API.
"""

from __future__ import print_function
import argparse
import getpass
import logging
import os
import sys

from oslo_utils import encodeutils
from oslo_utils import importutils

import harborclient
from harborclient import api_versions
from harborclient import client
from harborclient import exceptions as exc
from harborclient import utils

DEFAULT_API_VERSION = "2.0"
DEFAULT_MAJOR_OS_COMPUTE_API_VERSION = "2.0"

logger = logging.getLogger(__name__)


class HarborClientArgumentParser(argparse.ArgumentParser):
    def __init__(self, *args, **kwargs):
        super(HarborClientArgumentParser, self).__init__(*args, **kwargs)

    def error(self, message):
        """error(message: string)

        Prints a usage message incorporating the message to stderr and
        exits.
        """
        self.print_usage(sys.stderr)
        # FIXME(lzyeval): if changes occur in argparse.ArgParser._check_value
        choose_from = ' (choose from'
        progparts = self.prog.partition(' ')
        self.exit(2,
                  ("error: %(errmsg)s\nTry '%(mainp)s help %(subp)s'"
                   " for more information.\n") % {
                       'errmsg': message.split(choose_from)[0],
                       'mainp': progparts[0],
                       'subp': progparts[2]})

    def _get_option_tuples(self, option_string):
        """returns (action, option, value) candidates for an option prefix

        Returns [first candidate] if all candidates refers to current and
        deprecated forms of the same options parsing succeed.
        """
        option_tuples = (super(HarborClientArgumentParser, self)
                         ._get_option_tuples(option_string))
        if len(option_tuples) > 1:
            normalizeds = [
                option.replace('_', '-')
                for action, option, value in option_tuples
            ]
            if len(set(normalizeds)) == 1:
                return option_tuples[:1]
        return option_tuples


class HarborShell(object):
    times = []

    def _append_global_identity_args(self, parser, argv):
        # Register the CLI arguments that have moved to the session object.
        parser.set_defaults(os_username=utils.env('HARBOR_USERNAME'))
        parser.set_defaults(os_password=utils.env('HARBOR_PASSWORD'))
        parser.set_defaults(os_project=utils.env('HARBOR_PROJECT'))
        parser.set_defaults(os_baseurl=utils.env('HARBOR_URL'))

    def get_base_parser(self, argv):
        parser = HarborClientArgumentParser(
            prog='harbor',
            description=__doc__.strip(),
            epilog='See "harbor help COMMAND" '
            'for help on a specific command.',
            add_help=False,
            formatter_class=HarborHelpFormatter, )

        # Global arguments
        parser.add_argument(
            '-h',
            '--help',
            action='store_true',
            help=argparse.SUPPRESS, )

        parser.add_argument(
            '--debug',
            default=False,
            action='store_true',
            help="Print debugging output.")

        parser.add_argument(
            '--timings',
            default=False,
            action='store_true',
            help="Print call timing info.")

        parser.add_argument(
            '--version', action='version', version=harborclient.__version__)

        parser.add_argument(
            '--os-username',
            dest='os_username',
            metavar='<username>',
            help='Username')

        parser.add_argument(
            '--os-password',
            dest='os_password',
            metavar='<password>',
            help="User's password")

        parser.add_argument(
            '--os-project',
            dest='os_project',
            metavar='<project>',
            help="Project Id")

        parser.add_argument(
            '--timeout',
            metavar='<timeout>',
            help="Set request timeout (in seconds).")

        parser.add_argument(
            '--os-baseurl',
            metavar='<baseurl>',
            help='API base url')

        parser.add_argument(
            '--insecure',
            default=False,
            action='store_true',
            dest='insecure',
            help='Explicitly allow client to perform '
            '"insecure" TLS (https) requests. The '
            'server\'s certificate will not be verified '
            'against any certificate authorities. This '
            'option should be used with caution.')

        parser.add_argument(
            '--os-cacert',
            dest='os_cacert',
            metavar='<ca-certificate>',
            default=os.environ.get('OS_CACERT'),
            help='Specify a CA bundle file to use in '
            'verifying a TLS (https) server certificate. '
            'Defaults to env[OS_CACERT].')

        parser.add_argument(
            '--os-api-version',
            metavar='<api-version>',
            default=utils.env(
                'HARBOR_API_VERSION', default=DEFAULT_API_VERSION),
            help=('Accepts X, X.Y (where X is major and Y is minor part) or '
                  '"X.latest", defaults to env[HARBOR_API_VERSION].'))

        self._append_global_identity_args(parser, argv)

        return parser

    def get_subcommand_parser(self, version, do_help=False, argv=None):
        parser = self.get_base_parser(argv)

        self.subcommands = {}
        subparsers = parser.add_subparsers(metavar='<subcommand>')

        actions_module = importutils.import_module(
            "harborclient.v%s.shell" % version.ver_major)

        self._find_actions(subparsers, actions_module, version, do_help)
        self._find_actions(subparsers, self, version, do_help)
        self._add_bash_completion_subparser(subparsers)

        return parser

    def _add_bash_completion_subparser(self, subparsers):
        subparser = subparsers.add_parser(
            'bash_completion',
            add_help=False,
            formatter_class=HarborHelpFormatter)
        self.subcommands['bash_completion'] = subparser
        subparser.set_defaults(func=self.do_bash_completion)

    def _find_actions(self, subparsers, actions_module, version, do_help):
        msg = " (Supported by API versions '%(start)s' - '%(end)s')"
        for attr in (a for a in dir(actions_module) if a.startswith('do_')):
            # I prefer to be hyphen-separated instead of underscores.
            command = attr[3:].replace('_', '-')
            callback = getattr(actions_module, attr)
            desc = callback.__doc__ or ''
            if hasattr(callback, "versioned"):
                additional_msg = ""
                subs = api_versions.get_substitutions(
                    utils.get_function_name(callback))
                if do_help:
                    additional_msg = msg % {
                        'start': subs[0].start_version.get_string(),
                        'end': subs[-1].end_version.get_string()
                    }
                subs = [
                    versioned_method for versioned_method in subs
                    if version.matches(versioned_method.start_version,
                                       versioned_method.end_version)
                ]
                if subs:
                    # use the "latest" substitution
                    callback = subs[-1].func
                else:
                    # there is no proper versioned method
                    continue
                desc = callback.__doc__ or desc
                desc += additional_msg

            action_help = desc.strip()
            arguments = getattr(callback, 'arguments', [])

            subparser = subparsers.add_parser(
                command,
                help=action_help,
                description=desc,
                add_help=False,
                formatter_class=HarborHelpFormatter)
            subparser.add_argument(
                '-h',
                '--help',
                action='help',
                help=argparse.SUPPRESS, )
            self.subcommands[command] = subparser
            for (args, kwargs) in arguments:
                start_version = kwargs.get("start_version", None)
                if start_version:
                    start_version = api_versions.APIVersion(start_version)
                    end_version = kwargs.get("end_version", None)
                    if end_version:
                        end_version = api_versions.APIVersion(end_version)
                    else:
                        end_version = api_versions.APIVersion(
                            "%s.latest" % start_version.ver_major)
                    if do_help:
                        kwargs["help"] = kwargs.get("help", "") + (
                            msg % {
                                "start": start_version.get_string(),
                                "end": end_version.get_string()
                            })
                    if not version.matches(start_version, end_version):
                        continue
                kw = kwargs.copy()
                kw.pop("start_version", None)
                kw.pop("end_version", None)
                subparser.add_argument(*args, **kw)
            subparser.set_defaults(func=callback)

    def setup_debugging(self, debug):
        if not debug:
            return
        streamformat = "%(levelname)s (%(module)s:%(lineno)d) %(message)s"
        logging.basicConfig(level=logging.DEBUG, format=streamformat)
        logging.getLogger('iso8601').setLevel(logging.WARNING)

    def main(self, argv):
        # Parse args once to find version and debug settings
        parser = self.get_base_parser(argv)
        (args, args_list) = parser.parse_known_args(argv)
        self.setup_debugging(args.debug)
        do_help = ('help' in argv) or ('--help' in argv) or (
            '-h' in argv) or not argv

        # bash-completion should not require authentication
        if not args.os_api_version:
            api_version = api_versions.get_api_version(
                DEFAULT_MAJOR_OS_COMPUTE_API_VERSION)
        else:
            api_version = api_versions.get_api_version(args.os_api_version)

        os_username = args.os_username
        os_password = args.os_password
        os_project = args.os_project
        os_baseurl = args.os_baseurl
        subcommand_parser = self.get_subcommand_parser(
            api_version, do_help=do_help, argv=argv)
        self.parser = subcommand_parser

        if args.help or not argv:
            subcommand_parser.print_help()
            return 0

        args = subcommand_parser.parse_args(argv)

        # Short-circuit and deal with help right away.
        if args.func == self.do_help:
            self.do_help(args)
            return 0
        elif args.func == self.do_bash_completion:
            self.do_bash_completion(args)
            return 0
        insecure = args.insecure
        cacert = args.os_cacert
        if not os_baseurl:
            print(("ERROR (CommandError): You must provide harbor url via "
                   "either --os-baseurl or env[HARBOR_URL]."))
            return 1
        if not os_username:
            print(("ERROR (CommandError): You must provide username via "
                   "either --os-username or env[HARBOR_USERNAME]."))
            return 1
        if not os_project:
            print(("ERROR (CommandError): You must provide project via "
                   "either --os-project or env[HARBOR_PROJECT]."))
            return 1
        while not os_password:
            os_password = getpass.getpass("password: ")
        self.cs = client.Client(
            api_version,
            os_username,
            os_password,
            os_project,
            os_baseurl,
            timings=args.timings,
            http_log_debug=args.debug,
            insecure=insecure,
            cacert=cacert,
            timeout=args.timeout)
        try:
            self.cs.authenticate()
        except exc.Unauthorized:
            raise exc.CommandError("Invalid Harbor credentials.")
        except exc.AuthorizationFailure as e:
            raise exc.CommandError("Unable to authorize user '%s': %s"
                                   % (os_username, e))
        args.func(self.cs, args)
        if args.timings:
            self._dump_timings(self.times + self.cs.get_timings())

    def _dump_timings(self, timings):
        results = [{
            "url": url,
            "seconds": end - start
        } for url, start, end in timings]
        total = 0.0
        for tyme in results:
            total += tyme['seconds']
        results.append({"url": "Total", "seconds": total})
        utils.print_list(results, ["url", "seconds"], align='l')
        print("Total: %s seconds" % total)

    def do_bash_completion(self, _args):
        """Print bash completion

        Prints all of the commands and options to stdout so that the
        harbor.bash_completion script doesn't have to hard code them.
        """
        commands = list()
        options = list()
        for sc_str, sc in self.subcommands.items():
            commands.append(sc_str)
            for option in sc._optionals._option_string_actions.keys():
                options.append(option)

        options.extend(self.parser._option_string_actions.keys())
        print(' '.join(set(commands + options)))

    @utils.arg(
        'command',
        metavar='<subcommand>',
        nargs='?',
        help='Display help for <subcommand>.')
    def do_help(self, args):
        """Display help about this program or one of its subcommands."""
        if args.command:
            if args.command in self.subcommands:
                self.subcommands[args.command].print_help()
            else:
                raise exc.CommandError(
                    ("'%s' is not a valid subcommand") % args.command)
        else:
            self.parser.print_help()


# I'm picky about my shell help.
class HarborHelpFormatter(argparse.HelpFormatter):
    def __init__(self,
                 prog,
                 indent_increment=2,
                 max_help_position=32,
                 width=None):
        super(HarborHelpFormatter, self).__init__(prog, indent_increment,
                                                  max_help_position, width)

    def start_section(self, heading):
        # Title-case the headings
        heading = '%s%s' % (heading[0].upper(), heading[1:])
        super(HarborHelpFormatter, self).start_section(heading)


def main():
    try:
        argv = [encodeutils.safe_decode(a) for a in sys.argv[1:]]
        HarborShell().main(argv)
    except KeyboardInterrupt:
        print("... terminating harbor client", file=sys.stderr)
        sys.exit(130)
    except exc.CommandError as e:
        print("CommandError: %s" % e)
        sys.exit(127)


if __name__ == "__main__":
    main()
