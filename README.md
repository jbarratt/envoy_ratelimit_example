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

The external authorizer is a tiny bit of custom code, which implements the external authorizer gRPC spec.

The ratelimiter is the standard [lyft ratelimit](https://github.com/lyft/ratelimit), with the config in `ratelimit-data`.

The backend is a simple go http service.

There is also a simple vegeta script in `vegeta`. 

`make onepath` will try and issue 10rps against a single path in the mock backend. (It should get 2 rps with the ratelimiter config.)

`make twopaths` will issue 10rps alternating between two paths. Since it should get 2rps per path, you should see a total of 4 rps.

# Getting ready to build

* clone the repo
* `git submodule init`
* `git submodule update`


I had to make some manual tweaks to the `ratelimit` codebase to get it to build -- which may be operator error:

* `mkdir ratelimit/vendor` (the `Dockerfile` expects it to exist already)
* add a `COPY proto proto` to the `Dockerfile` with the rest of the `COPY` statements

Finally run:

* `docker-compose up`. The first one will take some time as it builds everything.



