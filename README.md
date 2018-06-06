# Hi!

This is a CF plugin that does zero downtime deployments.

# Why another plugin that does zero downtime deployments?

* First and foremost other plugins kind of hides what is going on under the hood.
 This plugin will always print out an execution plan with normal cf commands before doing its thing.
* The steps are different commands, thus allowing you to inject smoke tests or any other task you might want to do between each step.

# Ok, sounds good. So how does it work?

## Plugin
There are three plugins

### halfpipe-push

This simply deploys the application as `app-name-CANDIDATE` to a test route

### halfpipe-promote

* This binds all the routes from the manifest to the `app-name-CANDIDATE`
* Removes the test route from `app-name-CANDIDATE`
* renames `app-name-OLD` to `app-name-DELETE`
* renames `app-name` to `app-name-OLD` 
* renames `app-name-CANDIDATE` to `app-name`
* stops `app-name-OLD`

### halfpipe-delete

Simply deletes `app-name-DELETE`

## Concourse resource

```
resource_types:
- name: cf-resource
  type: docker-image
  source:
    repository: platformengineering/cf-resource
    tag: stable

- name: cf-resource
  type: cf-resource
  source:
    api: ((cloudfoundry.api-dev))
    org: my-org
    password: ((cloudfoundry.password))
    space: my-space
    username: ((cloudfoundry.username))
    
jobs:
- name: deploy-to-dev
  plan:
    - get: git
    - put: cf halfpipe-push
      resource: cf-resource
      params:
        appPath: git/target/distribution/artifact.zip
        command: halfpipe-push
        manifestPath: git/manifest.yml
        testDomain: some.random.domain.com
    - put: cf halfpipe-promote
      resource: cf-resource
      params:
        command: halfpipe-promote
        manifestPath: git/manifest.yml
        testDomain: some.random.domain.com
    - put: cf halfpipe-delete
      resource: cf-resource
      params:
        command: halfpipe-delete
        manifestPath: git/manifest.yml
```

# Cewl, how do I install the plugin?

```
$ go build go build cmd/plugin/plugin.go
$ cf install-plugin plugin
```

# Sample output
```
$ cf halfpipe-push -manifestPath manifest-dev.yml -appPath . -testDomain dev.cf.com -space dev
# CF plugin built from git revision '6ea00c8b3dc7c8251e259821a8fdd72916547462'
# Planned execution
#	* cf push halfpipe-example-nodejs-CANDIDATE -f manifest-dev.yml -p . -n halfpipe-example-nodejs-dev-CANDIDATE -d cf.dev.com

$ cf push halfpipe-example-nodejs-CANDIDATE -f manifest-dev.yml -p . -n halfpipe-example-nodejs-dev-CANDIDATE -d cf.dev.com
Using manifest file manifest-dev.yml

Creating app halfpipe-example-nodejs-CANDIDATE in org engineering-enablement / space dev as engineering-enablement@springernature.com...
OK
..
..

$ cf halfpipe-promote -manifestPath manifest-dev.yml -testDomain dev.cf.com -space dev
# CF plugin built from git revision '6ea00c8b3dc7c8251e259821a8fdd72916547462'
# Planned execution
#	* cf map-route halfpipe-example-nodejs-CANDIDATE dev.private.springernature.io -n halfpipe-example-nodejs
#	* cf unmap-route halfpipe-example-nodejs-CANDIDATE cf.dev.com -n halfpipe-example-nodejs-dev-CANDIDATE
#	* cf rename halfpipe-example-nodejs-OLD halfpipe-example-nodejs-DELETE
#	* cf rename halfpipe-example-nodejs halfpipe-example-nodejs-OLD
#	* cf stop halfpipe-example-nodejs-OLD
#	* cf rename halfpipe-example-nodejs-CANDIDATE halfpipe-example-nodejs

$ cf map-route halfpipe-example-nodejs-CANDIDATE dev.private.springernature.io -n halfpipe-example-nodejs
Creating route halfpipe-example-nodejs.dev.private.springernature.io for org engineering-enablement / space dev as engineering-enablement@springernature.com...
OK
Route halfpipe-example-nodejs.dev.private.springernature.io already exists
Adding route halfpipe-example-nodejs.dev.private.springernature.io to app halfpipe-example-nodejs-CANDIDATE in org engineering-enablement / space dev as engineering-enablement@springernature.com...
OK
...
...

$ cf halfpipe-cleanup -manifestPath manifest-dev.yml
# CF plugin built from git revision '6ea00c8b3dc7c8251e259821a8fdd72916547462'
# Planned execution
#	* cf delete halfpipe-example-nodejs-DELETE -f

$ cf delete halfpipe-example-nodejs-DELETE -f
Deleting app halfpipe-example-nodejs-DELETE in org engineering-enablement / space dev as engineering-enablement@springernature.com...
OK
```
