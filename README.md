# website-content-watcher

> An environment to execute puppet-master jobs on a regular basis, notifying users on change or with the current state.


## Usage

You can use the CLI either by installing the go package through running `go install github.com/Scalify/puppet-master-cli`
or by running the docker image:

```bash
docker run --rm -it scalify/website-content-watcher --help
```

you can find a straightforward example in the [example directory](example).

## Commands

### watch

Reads the given config and schedules cron jobs to watch the specified jobs. Example: 

**Assuming you are running from the root of this directory, using the provided `docker-compose.yml`, the `example` directory, and are running the self hosted example from [the official repo](https://github.com/Scalify/puppet-master/tree/master/examples/self_hosted)!**

```bash
docker-compose up -d redis mailcatcher
docker-compose run watcher
```

Then open [the mailcatcher interface](http://localhost:1080/) and see the mails incoming. :wink:

**Cleanup**: Press `CMD + c` to abort the watcher and run `docker-compose down` to remove the running containers and networks. 
## License

Copyright 2018 Scalify GmbH

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

		http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
