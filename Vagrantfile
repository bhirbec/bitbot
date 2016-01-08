Vagrant.configure("2") do |config|
    config.vm.box = "ubuntu/trusty64"
    config.vm.box_url = "https://atlas.hashicorp.com/ubuntu/boxes/trusty64/versions/20151218.0.0/providers/virtualbox.box"
    config.vm.provider :virtualbox do |vb| vb.name = "bitbot" end
    config.vm.synced_folder ".", "/vagrant"
    config.vm.network :forwarded_port, guest: 3306, host: 3306
    config.vm.network :forwarded_port, guest: 8080, host: 8080

	config.vm.provision "ansible" do |ansible|
		ansible.playbook = "ansible/setup.yaml"
	end
end
