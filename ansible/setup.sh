# TODO: check --maintenance-policy
# TODO: check --scopes

gcloud compute instances create recorder \
    --zone "us-central1-b" \
    --machine-type "f1-micro" \
    --maintenance-policy "MIGRATE" \
    --image "https://www.googleapis.com/compute/v1/projects/ubuntu-os-cloud/global/images/ubuntu-1404-trusty-v20150805" \
    --boot-disk-size "10" \
    --boot-disk-type "pd-standard" \
    --tags "http-server" "https-server";

# make sure we can ssh the instance
while [ "$(gcloud compute ssh recorder --command="echo ok")" != "ok" ]; do
    echo "Cannot SSH the instance. Retrying in 1 sec...";
    sleep 1;
done

ansible-playbook ansible/setup.yaml -i ansible/gce_hosts
