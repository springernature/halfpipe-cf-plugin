team: engineering-enablement

tasks:
- type: run
  script: ./test.sh
  docker:
    image: golang:1.10.0-alpine3.7
  vars:
    RUNNING_IN_CI: true # Used in test script to setup correct env
- type: docker-push
  username: platformengineering
  password: ((docker-hub-pe.password))
  image: platformengineering/halfpipe-cf-resource