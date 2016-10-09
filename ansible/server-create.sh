gcloud config set project the-bitbot-project;
gcloud config set compute/zone us-central1-b;

echo "Creating disk..."
gcloud compute disks create "bitbot-data" \
   --size "100" \
   --zone "us-central1-b" \
   --type "pd-standard";

# TODO: check --maintenance-policy
# TODO: check --scopes
echo "Creating instance..."
gcloud compute instances create bitbot-1 \
    --zone "us-central1-b" \
    --machine-type "f1-micro" \
    --maintenance-policy "MIGRATE" \
    --image "/ubuntu-os-cloud/ubuntu-1404-trusty-v20160919" \
    --boot-disk-size "10" \
    --boot-disk-type "pd-standard" \
    --disk name="bitbot-data",device-name="bitbot-data" \
    --tags "http-server,https-server";

# make sure we can ssh the instance
while [ "$(gcloud compute ssh bitbot-1 --zone "us-central1-b" --command="echo ok")" != "ok" ]; do
    echo "Cannot SSH the instance. Retrying in 1 sec...";
    sleep 1;
done

gcloud compute config-ssh;

echo "Formatting & Mounting disk..."
gcloud compute ssh bitbot-1 --zone "us-central1-b" --command '
    sudo mkfs.ext4 -F -E lazy_itable_init=0,lazy_journal_init=0,discard /dev/disk/by-id/google-bitbot-data;
    sudo mkdir -p /mnt/bitbot-data;
    sudo mount -o discard,defaults /dev/disk/by-id/google-bitbot-data /mnt/bitbot-data;
    sudo chmod a+w /mnt/bitbot-data;
    echo UUID=`sudo blkid -s UUID -o value /dev/disk/by-id/google-bitbot-data` /mnt/bitbot-data ext4 discard,defaults,[NOFAIL] 0 2 | sudo tee -a /etc/fstab;
';

ansible-playbook ansible/provision.yaml -i ansible/gce_hosts
