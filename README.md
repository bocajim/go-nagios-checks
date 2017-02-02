# go-nagios-checks

Key Features
------------
* Checks graphite metrics for errors
* Checks AWS cloudwatch for errors

Usage
-----
```
bin/nagios-checks -m=aws-cw -ak=<access_id> -sk=<secret_key> -rg=us-east-1 -ns=AWS/EBS -mn=BurstBalance -w=2440 -d=VolumeId -v=vol-8764f72f -wc="<50" -cc="<20"
````