# Envoy Custom Ratelimiter Via Authorizer

The premise is to

* Use the authorizer to inject a custom header
* Pass that header's value to the rate limiter
* Use it for rate limiting


# The moving parts

The call flow is

```
- envoy on localhost port 8010
- checks with the external authorizer (extauth)
- that external authorizer sets a header
- envoy then checks with the ratelimiter, which persists it's state in redis
- if the check passes, it passes it to the backend, which returns a response which includes all the headers which were passed to it
```

The external authorizer is a tiny bit of custom code, which implements the external authorizer GRPC spec.

The ratelimiter is the standard [lyft ratelimit](https://github.com/lyft/ratelimit), with the config in `ratelimit-data`.

The backend is a simple go http service.

There is also a simple vegeta script in `vegeta`, if you run `make loadtest` it will try and issue a sample workload so you can verify rate limiting is working as intended.

# Getting ready to build

* clone the repo
* `git submodule init`
* `git submodule update`
* `docker-compose up`. The first one will take some time as it builds everything.


# Submodule care and feeding

The submodule work is in a branch called `buildtweaks`.  (`git checkout buildtweaks`)

For periodically pulling in updates

* `git submodule update --remote --rebase ratelimit`

When pushing changes to it make sure to

* `git push --recurse-submodules=on-demand`


