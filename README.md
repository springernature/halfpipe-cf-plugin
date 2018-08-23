# Hi!

This is a CF plugin that does zero downtime deployments.

# Why another plugin that does zero downtime deployments?

* First and foremost other plugins kind of hides what is going on under the hood.
 This plugin will always print out an execution plan with normal cf commands before doing its thing.
* The steps are different commands, thus allowing you to inject smoke tests or any other task you might want to do between each step.

# Ok, sounds good. So how does it work?

There are three plugins and given a manifest like

```
applications:
- name: app-name
  memory: 50MB
  instances: 1
  routes:
  - route: route1.domain.com
  - route: fullDomain.com
```

## halfpipe-push

This simply deploys the application as `app-name-CANDIDATE` to a test route `app-name-{SPACE}-CANDIDATE.{DOMAIN}`

## halfpipe-promote

* This binds all the routes from the manifest to the `app-name-CANDIDATE`
* Removes the test route from `app-name-CANDIDATE`
* renames `app-name-OLD` to `app-name-DELETE`
* renames `app-name` to `app-name-OLD` 
* renames `app-name-CANDIDATE` to `app-name`
* stops `app-name-OLD`

## halfpipe-cleanup

Simply deletes the app `app-name-DELETE`

# Ok, so a lot of talk about routes. What if I have a worker app?
Just put `no-route: true` in the manifest!

# Sample output

Given a manifest like 

```
$ cf halfpipe-push -manifestPath path/to/manifest-dev.yml -appPath path/to/app -testDomain dev.cf.com
# CF plugin built from git revision '73a073793cd3bbce60428423a53e7544685dfce3'
# Planned execution
#	* cf push my-app-CANDIDATE -f path/to/manifest-dev.yml -p path/to/app --no-route --no-start
#	* cf map-route my-app-CANDIDATE dev.cf.com -n my-app-dev-CANDIDATE
#	* cf start my-app-CANDIDATE

$ cf push my-app-CANDIDATE -f path/to/manifest-dev.yml -p path/to/app --no-route --no-start
Using manifest file path/to/manifest-dev.yml

Creating app my-app-CANDIDATE in org engineering-enablement / space dev as engineering-enablement@springernature.com...
OK

App my-app-CANDIDATE is a worker, skipping route creation
Uploading my-app-CANDIDATE...
Uploading app files from: /tmp/build/put/git/nodejs
Uploading 7.7K, 11 files
Done uploading               
OK

$ cf map-route my-app-CANDIDATE dev.cf.com -n my-app-dev-CANDIDATE
Creating route my-app-dev-CANDIDATE.dev.cf.com for org engineering-enablement / space dev as engineering-enablement@springernature.com...
OK
Adding route my-app-dev-CANDIDATE.dev.cf.com to app my-app-CANDIDATE in org engineering-enablement / space dev as engineering-enablement@springernature.com...
OK

$ cf start my-app-CANDIDATE
Starting app my-app-CANDIDATE in org engineering-enablement / space dev as engineering-enablement@springernature.com...
Creating container
Downloading app package...
Downloaded app package (7.5K)
Successfully created container
....
App started


OK

App my-app-CANDIDATE was started using this command `node app.js`

Showing health and status for app my-app-CANDIDATE in org engineering-enablement / space dev as engineering-enablement@springernature.com...
OK

requested state: started
instances: 1/1
usage: 50M x 1 instances
urls: my-app-dev-CANDIDATE.dev.cf.com
last uploaded: Wed Aug 8 13:15:47 UTC 2018
stack: cflinuxfs2
buildpack: https://github.com/cloudfoundry/nodejs-buildpack#v1.6.17

     state     since                    cpu    memory     disk      details
#0   running   2018-08-08 01:16:24 PM   0.0%   0 of 50M   0 of 1G

```

```
$ cf halfpipe-promote -manifestPath path/to/manifest-dev.yml -testDomain dev.cf.com
# CF plugin built from git revision '73a073793cd3bbce60428423a53e7544685dfce3'
# Planned execution
#	* cf map-route my-app-CANDIDATE dev.cf.com -n my-app
#	* cf map-route my-app-CANDIDATE dev.cf.com -n some-other-host --path nodejs
#	* cf unmap-route my-app-CANDIDATE dev.cf.com -n my-app-dev-CANDIDATE
#	* cf rename my-app-OLD my-app-DELETE
#	* cf rename my-app my-app-OLD
#	* cf stop my-app-OLD
#	* cf rename my-app-CANDIDATE my-app

$ cf map-route my-app-CANDIDATE dev.cf.com -n my-app
Creating route my-app.dev.cf.com for org engineering-enablement / space dev as engineering-enablement@springernature.com...
OK
Adding route my-app.dev.cf.com to app my-app-CANDIDATE in org engineering-enablement / space dev as engineering-enablement@springernature.com...
OK

$ cf map-route my-app-CANDIDATE dev.cf.com -n some-other-host --path nodejs
Creating route some-other-host.dev.cf.com/nodejs for org engineering-enablement / space dev as engineering-enablement@springernature.com...
OK
Adding route some-other-host.dev.cf.com/nodejs to app my-app-CANDIDATE in org engineering-enablement / space dev as engineering-enablement@springernature.com...
OK

$ cf unmap-route my-app-CANDIDATE dev.cf.com -n my-app-dev-CANDIDATE
Removing route my-app-dev-CANDIDATE.dev.cf.com from app my-app-CANDIDATE in org engineering-enablement / space dev as engineering-enablement@springernature.com...
OK

$ cf rename my-app-OLD my-app-DELETE
Renaming app my-app-OLD to my-app-DELETE in org engineering-enablement / space dev as engineering-enablement@springernature.com...
OK

$ cf rename my-app my-app-OLD
Renaming app my-app to my-app-OLD in org engineering-enablement / space dev as engineering-enablement@springernature.com...
OK

$ cf stop my-app-OLD
Stopping app my-app-OLD in org engineering-enablement / space dev as engineering-enablement@springernature.com...
OK

$ cf rename my-app-CANDIDATE my-app
Renaming app my-app-CANDIDATE to my-app in org engineering-enablement / space dev as engineering-enablement@springernature.com...
OK
```

```
$ cf halfpipe-cleanup -manifestPath path/to/manifest-dev.yml
# CF plugin built from git revision '73a073793cd3bbce60428423a53e7544685dfce3'
# Planned execution
#	* cf delete my-app-DELETE -f

$ cf delete my-app-DELETE -f
Deleting app my-app-DELETE in org engineering-enablement / space dev as engineering-enablement@springernature.com...
OK
```



