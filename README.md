# angi

Quick implementation of controller requirements for interview process

## Description

Controller implementing the CRD provided in the PDF. It implements the creation and deletion of resources. Doesn't apply updates.
Testing revolves around Postman in the test/ folder

### Running and Testing

1. Bring up Minikube

```sh
minikube start
```

2. Make docker daemon the same as minikube

```sh
eval $(minikube docker-env)
```

3. In the top directory run the following commands. As long as the image is returned it should be available to minikube

```sh
make docker-build
docker images | grep controller
```

4. Deploy CRD and Controller:

```sh
make deploy
```

5. Apply the custom resources:

```sh
kubectl apply -f config/samples/group_v1alpha1_myappresource.yaml
```

6. Port forward the deployment

```sh
kubectl port-forward deployment/podinfo-myappresource-sample 8080:9898
```

7. Test it out
   At this point you can open a browser and go to "localhost:8080" and view the webpage that should include the custom message and color chosen
   Depending on if Redis is enabled you can also test the endpoint via Postman. You will want to import the Postman json from the tests/ folder.

### License

Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
