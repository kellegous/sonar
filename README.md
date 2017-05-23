# Sonar

## Makeshift Getting Started

I'm working on a proper README, but here's how you can use this thing right now with the docker image I created.

 1. `mkdir data`
 2. Create the file `data/sonar.toml` with the following contents:
```
addr = ":7699"
data-path = "data"
samples-per-period = 10
sample-period = "1min"

[[hosts]]
ip = "8.8.8.8"
name = "Google"

[[hosts]]
ip = "77.88.8.8"
name = "Russia (Yandex)"

[[hosts]]
ip = "91.239.100.100"
name = "Copenhagen, Denmark"

[[hosts]]
ip = "125.227.80.43"
name = "Beijing, China"
```
 3. `docker run -ti --rm -v $(pwd)/data:/data -p 7699:7699 kellegous/sonar`