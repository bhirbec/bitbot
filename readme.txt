GCE
===

Create a VM instance and disk
-----------------------------
$ ./ansible/setup.sh

Deployment
----------

- Install Ansible
$ git clone git://github.com/ansible/ansible.git --recursive
$ cd ./ansible
$ source ./hacking/env-setup
$ git pull --rebase
$ git submodule update --init --recursive
$ sudo make install

- run ansible-playbook
$ ansible-playbook ansible/deploy.yaml -i ansible/gce_hosts
