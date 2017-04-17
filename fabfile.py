#! usr/bin/python  
#pylint: disable=all 
from __future__ import with_statement
from fabric.api import local, settings, abort, run, cd, env, put,lcd

env.hosts = ['tonnn@115.159.101.234']
env.key_filename = '~/.ssh/id_elastic'

BIN_NAME = "gamehealthysrv"
SERVICE_NAME = "gamehealthysrv.service"

def build(os='darwin', arch='amd64', binary=BIN_NAME):
    with lcd('./cmd'):
        targetPath = "/".join(['../build', os, arch])
        target = targetPath + '/' + binary
        local("mkdir -p " + targetPath)
        local('GOOS='+os+' GOARCH='+arch+' go build -o '+target)


service = '/usr/lib/systemd/system/' + SERVICE_NAME

targetDir = {
    "bin_dir": '/data/'+BIN_NAME+'/bin',
    "conf_dir": '/data/'+BIN_NAME+'/configs',
}

user = "gamesms"
group = "gamesms"

def preDeploy():
    run('sudo adduser -U ' + user, warn_only=True)
    for k in targetDir.keys():
        run('sudo mkdir -p ' + targetDir[k])
        run('sudo chown '+user+':'+group+' '+targetDir[k])
   

def deploy():
    preDeploy()
    with cd(targetDir["bin_dir"]):
        put('./build/linux/amd64/' + BIN_NAME, BIN_NAME, use_sudo=True)
        run('sudo chmod u+x ' + BIN_NAME)
        run('sudo chown ' + user + ':' + group + ' ' + BIN_NAME)

    put('./configs/remote/gamesmssrv.json', targetDir["conf_dir"], use_sudo=True)
    put('./' + SERVICE_NAME, service, use_sudo=True)
    run('sudo systemctl enable ' + SERVICE_NAME)


def start():
    run('sudo systemctl daemon-reload')
    run('sudo systemctl restart ' + SERVICE_NAME)


def stop():
    run('sudo systemctl stop ' + SERVICE_NAME)


def status():
    run('sudo systemctl status ' + SERVICE_NAME)

def log(lines=20, tail=True):
    argTail = '-f'
    if not tail:
        argTail = ''
    run('sudo journalctl '+argTail+' -l -n '+str(lines)+' -u ' + SERVICE_NAME)