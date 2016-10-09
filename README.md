Fun side project that fetches cryptocurrency orderbooks, computes arbitrage opportunities and show
cool D3 charts.

Technical stack includes:  

  * [Golang](https://golang.org/)
  * [MySQL 5.7](https://dev.mysql.com/doc/refman/5.7/en/json.html) (with JSON support)
  * [React](http://facebook.github.io/react/)
  * [React Router](https://github.com/reactjs/react-router)
  * [Material UI](http://www.material-ui.com/)
  * [Browserify](http://browserify.org/)
  * [Vagrant](https://www.vagrantup.com/)
  * [Ansible](http://docs.ansible.com/)
  * [Google Compute Engine](https://cloud.google.com/compute/)
 
# Install

## Vagrant Install

Install [VirtualBox](https://www.virtualbox.org/wiki/Downloads) and [Vagrant](https://www.vagrantup.com/).

Create the Vagrant machine:  
`$ vagrant up`

Connect to the VM:  
`$ vagrant ssh`

Create MySQL database:  
`$ mysql -u bitbot -p bitbot < db/init.sql`  
password

## Start the Services

You first need to compile the Go code and build the JavaScript application:  
`$ cd /vagrant`  
`$ make`  

Now you can start the `record` service which fetches Bitcoin orderbooks from several exchangers:

`$ bin/record`

Start the `web` service that powers the UI and provides the API (you will need another SSH session):

`$ bin/web -b 0.0.0.0:8080`

Open your browser and point it at [localhost:8080](http://localhost:8080)

**Note**: you will need to run make and restart the services each time you make a change to the code.

# Deploy the Code on GCE

## Install Ansible

`$ git clone git://github.com/ansible/ansible.git --recursive`  
`$ cd ./ansible`  
`$ source ./hacking/env-setup`  
`$ git pull --rebase`  
`$ git submodule update --init --recursive`  
`$ sudo make install`

## Create a VM Instance and Disk

`$ ./ansible/server-create.sh`

## Deployment

Run ansible-playbook:

`$ ansible-playbook ansible/deploy.yaml -i ansible/gce_hosts`  
