Kraken exchange sends email in order to confirm withdraw requests. These email contain a link that the user 
needs to click on in order to validate the withdrawn. The purpose of this service is to automate this process.

First you will need to authorize the application 

$ withdrawn-confirm -api-keys <path> authorize

Follow the instructions printed on the console. This operation needs to be ran once on the server (for now this 
is done manually). 

After this has be completed you will need to restart the service:
$ sudo service withdraw-confirm restart

