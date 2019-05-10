from shutil import copyfile
import os
import sys
import argparse

if sys.version_info[:3][0] == 2:
    import ConfigParser as ConfigParser
    import StringIO as StringIO

if sys.version_info[:3][0] == 3:
    import configparser as ConfigParser
    import io as StringIO

RC_VALIDATE = 101
RC_UP = 102
RC_DOWN = 103
RC_BACKUP = 104
RC_RESTORE = 105
RC_UNKNOWN_TYPE = 106
RC_GEN = 110

class DBMigrator():

    def __init__(self, target):
        self.target = target
        self.script = "./db/run.sh"

    def backup(self):
        return run_cmd(self.script + " backup") == 0

    def restore(self):
        return run_cmd(self.script + " restore") == 0

    def up(self):
        cmd = self.script + " up"
        if self.target != '':
            cmd = cmd + " " + self.target
        return run_cmd(cmd) == 0

    def validate(self):
        return run_cmd(self.script + " test") == 0

class CfgMigrator():

    def __init__(self, target, output):
        cfg_dir = "/harbor-migration/harbor-cfg"
        cfg_out_dir = "/harbor-migration/harbor-cfg-out"

        self.target = target
        self.cfg_path = self.__config_filepath(cfg_dir)
        if not self.cfg_path:
            self.cfg_path = os.path.join(cfg_dir, "harbor.cfg")

        if output:
            self.output = output
        elif self.__config_filepath(cfg_out_dir):
            self.output = self.__config_filepath(cfg_out_dir)
        elif os.path.isdir(cfg_out_dir):
            self.output = os.path.join(cfg_out_dir, os.path.basename(self.cfg_path))
        else:
            self.output = ""

        self.backup_path = "/harbor-migration/backup"
        self.restore_src = self.__config_filepath(self.backup_path)
        self.restore_tgt = os.path.join(os.path.dirname(self.cfg_path), os.path.basename(self.restore_src))

    @staticmethod
    def __config_filepath(d):
        if os.path.isfile(os.path.join(d, "harbor.yml")):
            return os.path.join(d, "harbor.yml")
        elif os.path.isfile(os.path.join(d, "harbor.cfg")):
            return os.path.join(d, "harbor.cfg")
        return ""

    def backup(self):
        try:
            copyfile(self.cfg_path, os.path.join(self.backup_path, os.path.basename(self.cfg_path)))
            print ("Success to backup harbor.cfg.")
            return True
        except Exception as e:
            print ("Back up error: %s" % str(e))
            return False 

    def restore(self):
        if not self.restore_src:
            print("unable to locate harbor config file in directory: %s" % self.backup_path)
            return False

        try:
            copyfile(self.restore_src, self.restore_tgt)
            print ("Success to restore harbor.cfg.")
            return True
        except Exception as e:
            print ("Restore error: %s" % str(e))
            return False

    def up(self):
        if not os.path.exists(self.cfg_path):
            print ("Skip cfg up as no harbor.cfg in the path.")
            return True

        if self.output and os.path.isdir(self.output):
            cmd = "python ./cfg/run.py --input " + self.cfg_path + " --output " + os.path.join(self.output, os.path.basename(self.cfg_path))
        elif self.output and os.path.isfile(self.output):
            cmd = "python ./cfg/run.py --input " + self.cfg_path + " --output " + self.output
        else:
            print ("The path of the migrated harbor.cfg is not set, the input file will be overwritten.")
            cmd = "python ./cfg/run.py --input " + self.cfg_path

        if self.target != '':
            cmd = cmd + " --target " + self.target
        print("Command for config file migration: %s" % cmd)
        return run_cmd(cmd) == 0

    def validate(self):
        if not os.path.exists(self.cfg_path):
            print ("Unable to locate the input harbor configuration file: %s, please check." % self.cfg_path)
            return False
        return True

class Parameters(object):
    def __init__(self):    
        self.db_user = os.getenv('DB_USR', '')
        self.db_pwd = os.getenv('DB_PWD', '')
        self.skip_confirm = os.getenv('SKIP_CONFIRM', 'n')
        self.output = False
        self.is_migrate_db = True
        self.is_migrate_cfg = True
        self.target_version = ''
        self.action = ''
        self.init_from_input()

    def is_action(self, action):
        if action == "test" or action == "backup" or action == "restore" or action == "up":
            return True     
        else:
            return False   

    def parse_input(self):
        argv_len = len(sys.argv[1:])
        last_argv = sys.argv[argv_len:][0]
        if not self.is_action(last_argv):
            print ("Fail to parse input: the last parameter should in test:up:restore:backup")
            sys.exit(RC_GEN) 

        if last_argv == 'up':
            if self.skip_confirm != 'y':
                if not pass_skip_confirm():
                    sys.exit(RC_GEN) 
        
        if argv_len == 1:
            return (True, True, '', False, last_argv)

        parser = argparse.ArgumentParser(description='migrator of harbor') 
        parser.add_argument('--db', action="store_true", dest='is_migrate_db', required=False, default=False, help='The flag to upgrade db.')
        parser.add_argument('--cfg', action="store_true", dest='is_migrate_cfg', required=False, default=False, help='The flag to upgrede cfg.')
        parser.add_argument('--version', action="store", dest='target_version', required=False, default='', help='The target version that the harbor will be migrated to.')         
        parser.add_argument('--output', action="store_true", dest='output', required=False, default=False, help='The path of the migrated harbor.cfg, if not set the input file will be overwritten.')         
    
        args = parser.parse_args(sys.argv[1:argv_len])
        args.action = last_argv
        return (args.is_migrate_db, args.is_migrate_cfg, args.target_version, args.output, args.action)

    def init_from_input(self):
        (self.is_migrate_db, self.is_migrate_cfg, self.target_version, self.output, self.action) = self.parse_input()

def run_cmd(cmd):
    return os.system(cmd)

def pass_skip_confirm():
    valid = {"yes": True, "y": True, "ye": True, "no": False, "n": False}
    message = "Please backup before upgrade, \nEnter y to continue updating or n to abort: "
    while True:
        sys.stdout.write(message)
        choice = raw_input().lower()
        if choice == '':
            return False
        elif choice in valid:
            return valid[choice]
        else:
            sys.stdout.write("Please respond with 'yes' or 'no' "
                             "(or 'y' or 'n').\n")

def main():
    commandline_input = Parameters()

    db_migrator = DBMigrator(commandline_input.target_version)
    cfg_migrator = CfgMigrator(commandline_input.target_version, commandline_input.output)

    try:
        # test
        if commandline_input.action == "test":
            if commandline_input.is_migrate_db:
                if not db_migrator.validate():
                    print ("Fail to validate: please make sure your DB auth is correct.")
                    sys.exit(RC_VALIDATE)                    

            if commandline_input.is_migrate_cfg:
                if not cfg_migrator.validate():                 
                    print ("Fail to validate: please make sure your cfg path is correct.")
                    sys.exit(RC_VALIDATE) 
        
        # backup
        elif commandline_input.action == "backup":
            if commandline_input.is_migrate_db:
                if not db_migrator.backup():
                    sys.exit(RC_BACKUP)                    

            if commandline_input.is_migrate_cfg:
                if not cfg_migrator.backup():                 
                    sys.exit(RC_BACKUP)         
        
        # up
        elif commandline_input.action == "up":
            if commandline_input.is_migrate_db:
                if not db_migrator.up():
                    sys.exit(RC_UP)                    

            if commandline_input.is_migrate_cfg:
                if not cfg_migrator.up():                 
                    sys.exit(RC_UP)
        
        # restore
        elif commandline_input.action == "restore":
            if commandline_input.is_migrate_db:
                if not db_migrator.restore():
                    sys.exit(RC_RESTORE)                    

            if commandline_input.is_migrate_cfg:
                if not cfg_migrator.restore():                 
                    sys.exit(RC_RESTORE)
        
        else:
            print ("Unknow action type: " + str(commandline_input.action))
            sys.exit(RC_UNKNOWN_TYPE)
    except Exception as ex:
        print ("Migrator fail to execute, err: " + ex.message)
        sys.exit(RC_GEN)

if __name__ == '__main__':
    main()
