# Remove Instance & Disk
gcloud compute instances detach-disk bitbot-1 --disk bitbot-data --zone "us-central1-b" -q
gcloud compute instances delete bitbot-1 --zone us-central1-b -q
gcloud compute disks delete bitbot-data --zone "us-central1-b" -q
