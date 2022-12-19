# Varnish service controller

Project status: POC

#### What is it?

``vsvc`` is a lite Go application that runs in your k8s cluster. It scans all ``service``
objects and reads their annotations.
If any ``service`` is annotated with the ``varnish.backend`` annotation it will be picked up
by ``vsvc`` and will update Varnish's vcl accordingly.

#### How do I run it?

Apply the manifests found in the ``yaml`` folder within this repository.

Example:

```sh
$ kubectl create ns varnish
$ kubectl create ns nginx-demo
$ kubectl -n varnish apply -f yaml/varnish.yaml
$ kubectl -n nginx-demo apply -f yaml/nginx.yaml
```

In order to make the ``nginx-svc`` service available to Varnish run the following:

```sh
$ kubectl -n nginx-demo patch svc/nginx-svc -p '{"metadata":{"annotations":{"varnish.backend":"true"}}}'
```

``vsvc`` will notice the addition of the above annotation and generate the following vcl content:

```
vcl 4.1;

import directors;
import std;

backend nginx-svc {
    .host = "nginx-svc.nginx-demo.svc.cluster.local";
    .port = "8080";
}

sub vcl_recv {
    if (req.http.host == "nginx-svc") {
        set req.backend_hint = nginx-svc;
    }
}
```

For any other ``service`` object annotated in a similar fashion, vsvc will update the
above vcl accordingly. A ``varnishreload`` is triggered after every vcl update.

#### How do I test it?

If the setup above has been succesfull run the following:

```sh
$ kubectl -n varnish port-forward svc/varnish-svc 6081:6081
```

Note that ``6081`` is the service object port. Not to be confused with the port under
which Varnish actually runs in its pod (set to 80).

In another shell run the following request:

```sh
$ curl -H "Host: nginx-svc" 127.1:6081/
```

Check that the Varnish specific headers are found on the response.

#### How to remove a service from Varnish?

The only relation between a ``service`` object and Varnish is its annotation. In order
to have a ``service`` removed from Varnish's backend collection - we need to remove its annotation.
E.g:

```sh
$ kubectl -n nginx-demo annotate service nginx-svc varnish.backend-
```

#### How do I use the varnishtools?

All ``varnish*`` tools are available in the varnish container(s). It is possible to
exec into one such container and use these tools.

E.g:

```sh
$ kubectl -n varnish get po
$ kubectl -n varnish exec -it po/some-pod-name-here -c varnish -- /bin/sh
$ varnishlog -n /etc/varnish/work
```
