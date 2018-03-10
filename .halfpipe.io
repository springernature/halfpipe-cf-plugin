team: engineering-enablement

tasks:
- name: run
  script: ./test.sh
  docker:
    image: golang:1.10.0-alpine3.7
