# kubevirt-ip-helper-garbage-collector

This tool helps cleaning up orphaned vmnetcfg objects.

## Building the tool

Execute the go build command to build the tool:
```SH
go build -o kubevirt-ip-helper-garbage-collector .
```

## Usage

Make sure the kubeconfig of the KubeVirt cluster is used or point the KUBECONFIG environment variable to it, for example:
```SH
export KUBECONFIG=<PATH_TO_KUBECONFIG_FILE>
```

Just execute the garbage collector, it will be interactive:
```SH
./kubevirt-ip-helper-garbage-collector
```

# License

Copyright (c) 2023 Joey Loman <joey@binbash.org>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

[http://www.apache.org/licenses/LICENSE-2.0](http://www.apache.org/licenses/LICENSE-2.0)

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
