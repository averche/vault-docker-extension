1. Zombie extensions issue
   - the backing files are hard to find (not referenced in the error message)
   - even after removing the files, the docker images are started every time
2. Icon could not be switched (had to copy on top of `docker.svg`)
3. What is the canonical name of the extension?
   - one name specified in docker extension init
   - another name specified in "Hub repository" during the init step
4. A bit weird that `docker ps` does not show the containers for the running extensions by default
   - Shouldn't the goal be to interface with the extensions within the dev environment?
5. Great that I can use `docker-compose.yaml` but it's not very clear how it fits in:
   - `docker-compose.yaml` resides in the extenssion's docker image?
   - Couldn't figure out how to use a `build` tag within the compose file or where it runs
6. Is unix socket the only way to talk between React UI & the server?
   - would be nice to support standard HTTP listeners
   - couldn't find documentation on `metadata.json` `"exposes"/"socket"` elements