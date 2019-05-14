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

For now, it just returns a simple header:

```
Header: &core.HeaderValue{
	Key:   "x-ext-auth-ratelimit",
	Value: "magic",
},
```

In real life, this would probably be something like a user ID, or account ID, or the SHA of an API key ... whatever you want to rate limit on that you're aware of in your custom authorizer code. For this demo it's a simple static string. The important part is that it could be anything you want.

Envoy is configured (in `envoy.yaml`) to pass whatever value is set in that header, as well as the path the request was for, to the ratelimiter service.

```
rate_limits:
  - stage: 0
	actions:
	  - {request_headers: {header_name: "x-ext-auth-ratelimit", descriptor_key: "ratelimitkey"}}
	  - {request_headers: {header_name: ":path", descriptor_key: "path"}}
```

The ratelimiter is the standard [lyft ratelimit](https://github.com/lyft/ratelimit).

The config is buried down in `ratelimit-data/ratelimit/config/config.yaml` and is pretty simple:

```
domain: backend
descriptors:
  - key: ratelimitkey
    descriptors:
      - key: path
        rate_limit:
          requests_per_unit: 2
          unit: second
```

The domain is defined in the envoy config -- you can make it different for different parts of your service.

This config says to take the values that come with the `ratelimitkey` and `path` and build them into a joint key for rate limiting.

An example from the logs shows exactly how this works:

```
ratelimit_1  | time="2019-05-14T18:48:16Z" level=debug msg="cache key: backend_ratelimitkey_magic_path_/b_1557859696 current: 3"
ratelimit_1  | time="2019-05-14T18:48:16Z" level=debug msg="returning normal response"
```

So it's easy to imagine the string `magic` being replaced with something appropriate for your production environment.


The `backend` is a simple go http service. It prints the headers it gets to make it easy to see what headers are coming in with the request.

# Getting ready to build

* clone the repo
* `git submodule init`
* `git submodule update`


I had to make some manual tweaks to the `ratelimit` codebase to get it to build -- which may be operator error:

* `mkdir ratelimit/vendor` (the `Dockerfile` expects it to exist already)
* add a `COPY proto proto` to the `Dockerfile` with the rest of the `COPY` statements

Finally run:

* `docker-compose up`. The first one will take some time as it builds everything.

# Testing

You can ensure that the full stack is working with a simple curl:

```
$ curl -v localhost:8010
* Rebuilt URL to: localhost:8010/
*   Trying ::1...
* TCP_NODELAY set
* Connected to localhost (::1) port 8010 (#0)
> GET / HTTP/1.1
> Host: localhost:8010
> User-Agent: curl/7.54.0
> Accept: */*
> 
< HTTP/1.1 200 OK
< date: Tue, 14 May 2019 18:25:51 GMT
< content-length: 205
< content-type: text/plain; charset=utf-8
< x-envoy-upstream-service-time: 1
< server: envoy
< 
Oh, Hello!
Content-Length: 0
User-Agent: curl/7.54.0
Accept: */*
X-Forwarded-Proto: http
X-Request-Id: 46a043c4-2970-4a97-9c52-df220b938a01
X-Ext-Auth-Ratelimit: magic
* Connection #0 to host localhost left intact
X-Envoy-Expected-Rq-Timeout-Ms: 15000
```

There is also a simple vegeta script in `vegeta`. 
(If you don't have `vegeta` [you'll want to install it](https://github.com/tsenart/vegeta). It's my favorite load testing swiss army knife.)

`make onepath` will try and issue 10rps against a single path in the mock backend. (It should get 2 rps with the ratelimiter config.)

`make twopaths` will issue 10rps alternating between two paths. Since it should get 2rps per path, you should see a total of 4 rps.

```
$ make onepath
echo "GET http://localhost:8010/a" | vegeta attack -rate 10 -duration=15s | tee results.bin | vegeta report
Requests      [total, rate]            150, 10.07
Duration      [total, attack, wait]    14.913955151s, 14.900531s, 13.424151ms
Latencies     [mean, 50, 95, 99, max]  9.067402ms, 8.573452ms, 13.424151ms, 16.269111ms, 19.214047ms
Bytes In      [total, mean]            6882, 45.88
Bytes Out     [total, mean]            0, 0.00
Success       [ratio]                  20.67%
Status Codes  [code:count]             200:31  429:119  
Error Set:
429 Too Many Requests
```

So, attempting to do 10 queries per second, and getting only 20.67% success rate sounds about right.

```
$ make twopaths
echo "GET http://localhost:8010/a\nGET http://localhost:8010/b" | vegeta attack -rate 10 -duration=15s | tee results.bin | vegeta report
Requests      [total, rate]            150, 10.07
Duration      [total, attack, wait]    14.90981304s, 14.900302s, 9.51104ms
Latencies     [mean, 50, 95, 99, max]  9.409088ms, 8.732548ms, 13.591539ms, 16.279884ms, 22.797446ms
Bytes In      [total, mean]            13320, 88.80
Bytes Out     [total, mean]            0, 0.00
Success       [ratio]                  40.00%
Status Codes  [code:count]             200:60  429:90  
Error Set:
429 Too Many Requests
```

And since the config says we should get 2 requests/second for path, then having 40% of them succeed is ... perfect. Yay.
