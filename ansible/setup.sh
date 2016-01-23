# TODO: check --maintenance-policy
# TODO: check --scopes

gcloud compute disks create "bitbot-data" \
	--size "100" \
	--zone "us-central1-b" \
	--type "pd-standard";

gcloud compute instances create bitbot-1 \
    --zone "us-central1-b" \
    --machine-type "f1-micro" \
    --maintenance-policy "MIGRATE" \
    --image "https://www.googleapis.com/compute/v1/projects/ubuntu-os-cloud/global/images/ubuntu-1404-trusty-v20150805" \
    --boot-disk-size "10" \
    --boot-disk-type "pd-standard" \
    --disk name="bitbot-data",device-name="bitbot-data" \
    --tags "http-server,https-server";

# make sure we can ssh the instance
while [ "$(gcloud compute ssh bitbot-1 --command="echo ok")" != "ok" ]; do
    echo "Cannot SSH the instance. Retrying in 1 sec...";
    sleep 1;
done

# format disk
gcloud compute ssh bitbot-1 --command '
	sudo mkdir /media/bitbot-data;
	sudo /usr/share/google/safe_format_and_mount -m "mkfs.ext4 -F" /dev/disk/by-id/google-bitbot-data /media/bitbot-data;
';

# TODO: still need to edit ~/ssh/config :/
ansible-playbook ansible/setup.yaml -i ansible/gce_hosts -K;
