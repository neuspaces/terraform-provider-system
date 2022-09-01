# Alpine Linux test container

This container image runs Alpine Linux as a target for the acceptance tests of the Terraform provider system.

## Notes

- Container runs openrc as init process
- Redirects stdout/stderr from selected openrc services the container stdout/stderr

## Internal notes

- `Taskfile` implements tasks

## References

- Official Alpine Linux base image: https://hub.docker.com/_/alpine
