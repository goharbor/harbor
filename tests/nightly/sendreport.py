import json
import urllib2
from optparse import OptionParser
from jinja2 import Environment, FileSystemLoader
import os
import time
import re
import datetime
import email
import smtplib
import logging


class Parameters(object):
    def __init__(self):
        self.repo = ''
        self.branch = ''
        self.commmit = ''
        self.result = ''
        self.log = ''
        self.report_receiver = 'wangyan@vmware.com'
        self.from_address = 'wangyan@vmware.com'

        self.init_from_input()

    @staticmethod
    def parse_input():
        usage = "usage: %prog [options] <result set id>"
        parser = OptionParser(usage)
        parser.add_option("-r", "--repo", dest="repo", help="")
        parser.add_option("-b", "--branch", dest="branch", help="")
        parser.add_option("-c", "--commit", dest="commmit", help="")
        parser.add_option("-s", "--result", dest="result", help="")
        parser.add_option("-l", "--log", dest="log", help="")

        (options, args) = parser.parse_args()
        return (options.repo, options.branch, options.commmit, options.result, options.log)

    def init_from_input(self):
        (self.repo, self.branch, self.commmit, self.result, self.log) = Parameters.parse_input()


class EmailUtil:
    def __init__(self):
        pass

    @staticmethod
    def send_email(from_addr, to_addr, subject, body, times=1):
        try:
            mail = email.MIMEText.MIMEText(body, 'html')
            mail['From'] = from_addr
            mail['Subject'] = subject
            mail['To'] = to_addr
            smtp = smtplib.SMTP('mailhost.vmware.com')
            smtp.sendmail(mail['From'], mail['To'], mail.as_string())
            smtp.close()
        except Exception, e:
            logger.info(e)
            logger.info('send email fail, will try three times.')
            times += 1
            time.sleep(5)
            if times < 3:
                Utility.send_email(from_addr, to_addr, subject, body, times)

    @staticmethod
    def send_html_template(from_addr, to_addr, subject, html_obj):
        html_file = open('nightly-report.html', 'w')
        html_file.write(html_obj)
        html_file.close()
        EmailUtil.send_email(from_addr, to_addr, subject, html_obj)

class ReportRender:
    env = None
    template = None

    def __init__(self, index_file, commandline_input):
        self.env = Environment(loader=FileSystemLoader('tests/nightly'))
        self.template = self.env.get_template(index_file)
        self.commandline_input = commandline_input

    def render(self):
        return self.template.render(
            repo=self.commandline_input.repo,
            branch=self.commandline_input.branch,
            commit=self.commandline_input.commmit,
            result=self.commandline_input.result,
            log=self.commandline_input.log)

def main():
    commandline_input = Parameters()
    try:
        report_render = ReportRender('nightly-report-temp.html', commandline_input)
        report_html_obj = report_render.render()
        EmailUtil.send_html_template(commandline_input.from_address, commandline_input.report_receiver,
                               "Harbor nightly results", report_html_obj)
    except Exception, e:
        print str(e)

if __name__ == '__main__':
    main()
