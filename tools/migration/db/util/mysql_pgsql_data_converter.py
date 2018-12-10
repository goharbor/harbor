#!/usr/bin/env python

import re
import sys
import os
import time
import subprocess

def convert_registry_db(mysql_dump_file, pgsql_dump_file):
    mysql_dump = open(mysql_dump_file)
    pgsql_dump = open(pgsql_dump_file, "w")
    insert_lines = []

    for i, line in enumerate(mysql_dump):
        
        line = line.decode("utf8").strip()

        # catch insert
        if line.startswith("INSERT INTO"):
            # pgsql doesn't support user as a table name, change it to harbor_user.
            if line.startswith('INSERT INTO "user"'):
                insert_lines.append(line.replace('INSERT INTO "user"', 'INSERT INTO "harbor_user"'))
            # pgsql doesn't support upper-case as a column name, change it to lower-case.
            elif line.find('INSERT INTO "access_log" ("log_id", "username", "project_id", "repo_name", "repo_tag", "GUID", "operation", "op_time")') != -1:
                line = line.replace('INSERT INTO "access_log" ("log_id", "username", "project_id", "repo_name", "repo_tag", "GUID", "operation", "op_time")', 
                                    'INSERT INTO "access_log" ("log_id", "username", "project_id", "repo_name", "repo_tag", "guid", "operation", "op_time")')
                insert_lines.append(line)
                continue
            # pgsql doesn't support 0 as a time data, change it to the minimum value.
            elif line.find("0000-00-00 00:00:00") != -1:
                line = line.replace("0000-00-00 00:00:00", "0001-01-01 00:00:00")
                insert_lines.append(line)
                continue
            # mysqldump generates dumps in which strings are enclosed in quotes and quotes inside the string are escaped with a backslash 
            # like, {\"kind\":\"Manual\",\"schedule_param\":null}.
            # this is by design of mysql, see issue https://bugs.mysql.com/bug.php?id=65941
            # the data could be inserted into pgsql, but it will be failed on harbor api call.
            elif line.find('\\"') != -1:
                line = line.replace('\\"', '"')
                insert_lines.append(line)
                continue
            else:    
                insert_lines.append(line)
    
    write_database(pgsql_dump, "registry")
    write_insert(pgsql_dump, insert_lines)
    write_alter_table_bool(pgsql_dump, "harbor_user", "deleted")
    write_alter_table_bool(pgsql_dump, "harbor_user", "sysadmin_flag")
    write_alter_table_bool(pgsql_dump, "project", "deleted")
    write_alter_table_bool(pgsql_dump, "project_metadata", "deleted")
    write_alter_table_bool(pgsql_dump, "replication_policy", "enabled", "TRUE")
    write_alter_table_bool(pgsql_dump, "replication_policy", "replicate_deletion")
    write_alter_table_bool(pgsql_dump, "replication_policy", "deleted")
    write_alter_table_bool(pgsql_dump, "replication_target", "insecure")
    write_alter_table_bool(pgsql_dump, "replication_immediate_trigger", "on_push")
    write_alter_table_bool(pgsql_dump, "replication_immediate_trigger", "on_deletion")
    write_foreign_key(pgsql_dump)

    write_sequence(pgsql_dump, "harbor_user", "user_id")
    write_sequence(pgsql_dump, "project", "project_id")
    write_sequence(pgsql_dump, "project_member", "id")
    write_sequence(pgsql_dump, "project_metadata", "id")
    write_sequence(pgsql_dump, "user_group", "id")
    write_sequence(pgsql_dump, "access_log", "log_id")
    write_sequence(pgsql_dump, "repository", "repository_id")
    write_sequence(pgsql_dump, "replication_policy", "id")
    write_sequence(pgsql_dump, "replication_target", "id")
    write_sequence(pgsql_dump, "replication_immediate_trigger", "id")
    write_sequence(pgsql_dump, "img_scan_job", "id")
    write_sequence(pgsql_dump, "img_scan_overview", "id")
    write_sequence(pgsql_dump, "clair_vuln_timestamp", "id")
    write_sequence(pgsql_dump, "properties", "id")
    write_sequence(pgsql_dump, "harbor_label", "id")
    write_sequence(pgsql_dump, "harbor_resource_label", "id")
    write_sequence(pgsql_dump, "replication_job", "id")
    write_sequence(pgsql_dump, "role", "role_id")

def convert_notary_server_db(mysql_dump_file, pgsql_dump_file):
    mysql_dump = open(mysql_dump_file)
    pgsql_dump = open(pgsql_dump_file, "w")
    insert_lines = []

    for i, line in enumerate(mysql_dump):

        line = line.decode("utf8").strip()
        # catch insert
        if line.startswith("INSERT INTO"):
            if line.find("0000-00-00 00:00:00") != -1:
                line = line.replace("0000-00-00 00:00:00", "0001-01-01 00:00:00")
                insert_lines.append(line)
                continue
            else:    
                insert_lines.append(line)
    
    write_database(pgsql_dump, "notaryserver")
    write_insert(pgsql_dump, insert_lines)
    write_sequence(pgsql_dump, "tuf_files", "id")
    write_sequence(pgsql_dump, "changefeed", "id")

def convert_notary_signer_db(mysql_dump_file, pgsql_dump_file):
    mysql_dump = open(mysql_dump_file)
    pgsql_dump = open(pgsql_dump_file, "w")
    insert_lines = []

    for i, line in enumerate(mysql_dump):
        
        line = line.decode("utf8").strip()
        # catch insert
        if line.startswith("INSERT INTO"):
            if line.find("0000-00-00 00:00:00") != -1:
                line = line.replace("0000-00-00 00:00:00", "0001-01-01 00:00:00")
                insert_lines.append(line)
                continue
            else:    
                insert_lines.append(line)
    
    write_database(pgsql_dump, "notarysigner")
    write_insert(pgsql_dump, insert_lines)
    write_sequence(pgsql_dump, "private_keys", "id")

def write_database(pgsql_dump, db_name):
    pgsql_dump.write("\\c %s;\n" % db_name)

def write_table(pgsql_dump, table_lines):
    for item in table_lines:
        pgsql_dump.write("%s\n" % item)
        if item.startswith(');'):
            pgsql_dump.write('\n')
    pgsql_dump.write('\n')

def write_insert(pgsql_dump, insert_lines):
    for item in insert_lines:
        pgsql_dump.write("%s\n" % item.encode('utf-8'))

def write_foreign_key(pgsql_dump):
    pgsql_dump.write('\n')
    pgsql_dump.write("%s\n" % "ALTER TABLE \"project\" ADD CONSTRAINT \"project_ibfk_1\" FOREIGN KEY (\"owner_id\") REFERENCES \"harbor_user\" (\"user_id\");")
    pgsql_dump.write("%s\n" % "ALTER TABLE \"project_metadata\" ADD CONSTRAINT \"project_metadata_ibfk_1\" FOREIGN KEY (\"project_id\") REFERENCES \"project\" (\"project_id\");")

def write_alter_table_bool(pgsql_dump, table_name, table_columnn, default_value="FALSE"):
    pgsql_dump.write('\n')
    pgsql_dump.write("ALTER TABLE %s ALTER COLUMN %s DROP DEFAULT;\n" % (table_name, table_columnn))
    pgsql_dump.write("ALTER TABLE %s ALTER %s TYPE bool USING CASE WHEN %s=0 THEN FALSE ELSE TRUE END;\n" % (table_name, table_columnn, table_columnn))
    pgsql_dump.write("ALTER TABLE %s ALTER COLUMN %s SET DEFAULT %s;\n" % (table_name, table_columnn, default_value)) 

def write_sequence(pgsql_dump, table_name, table_columnn):
    pgsql_dump.write('\n')
    pgsql_dump.write("CREATE SEQUENCE IF NOT EXISTS %s_%s_seq;\n" % (table_name, table_columnn))
    pgsql_dump.write("SELECT setval('%s_%s_seq', max(%s)) FROM %s;\n" % (table_name, table_columnn, table_columnn, table_name))
    pgsql_dump.write("ALTER TABLE \"%s\" ALTER COLUMN \"%s\" SET DEFAULT nextval('%s_%s_seq');\n" % (table_name, table_columnn, table_name, table_columnn))

if __name__ == "__main__":
    if sys.argv[1].find("registry") != -1:
        convert_registry_db(sys.argv[1], sys.argv[2])
    elif sys.argv[1].find("notaryserver") != -1:
        convert_notary_server_db(sys.argv[1], sys.argv[2])
    elif sys.argv[1].find("notarysigner") != -1:
        convert_notary_signer_db(sys.argv[1], sys.argv[2])
    else:
        print ("Unsupport mysql dump file, %s" % sys.argv[1])
        sys.exit(1)