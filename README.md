# Distributed Key-Value Service Built on Chord
This is the Repo for term project of UCSD CSE223b. We implemented a load-balanced, easy-to-scale key-value service based on the [Chord protocol.](https://dl.acm.org/doi/pdf/10.1109/TNET.2002.808407)

## Get Started
To get started, first you should setup the environment variable for this go project. You should switch to the root path of this project, and run:
```shell
export GOPATH=$(pwd)
export PATH=$PATH:$GOPATH/bin
```

Then, you should compile the project and generate binary executives.
```shell
cd src/
make
```

You may now be able to use our server CLI to launch servers. First you should generate a backend configuration file using:
```shell
chord-mkrc -nnodes 5
```
where `-nnodes` specifies the number of backends (with random ip addresses as "localhost:$RAND") to generate. After that you may note a file `nodes.rc` containing 5 random backend ip addresses be generated under the current folder.

Then to launch the backend by
```shell
chord-node 0
```
This command will launch the 0th entry of backend configuration file. You can also launch different entries, e.g. `chord-node 1 2 4` will launch the 1st, 2nd and 4th entries. Note that this is a blocking call.

Once backend is launched, use client CLI to play with key-value service.
```shell
# set a to 1
chord-client set a 1
# get a
chord-client get a
# set a to 2 from the 0th backend. Make sure 0th backend has been launched
# Note that the key may not necessarily set on 0th backend. This index just serves as the access point of the backend group.
chord-client 0 set a 2
# get a from 2nd backend
chord-client 2 get a
```

## Visualize the Chord!
We support a visualizer for Chord, if you are interested in how Chord works. Currently we haven't integrated the code to master, so you must switch to the branch `demo` to enable this feature.
```shell
git branch demo
```

The visualizer depends on `ProtoBuf v3` of both Golang and Python. Therefore you may want to `make` to get required library.
```shell
cd src
make visual
```
This will run the shell script `src/visual/setup.sh` to automatically prepare everything for you. You can look at the shell script for more information.

Then you can launch the visualizer
```shell
cd src/visual
python visualize.py
```

Then you will see a GUI has been started with a dashed empty circle (that is the Chord ring!) on it. After that, you can use the client CLI to launch backends, do k-v operations, and so on.

Also, the code on `demo` branch is more advanced than master to support Chord node leave logic at the Chord proxier layer (not applicable to k-v storage layer, so may cause data loss). If you kill at most `r-1` nodes, where `r` is the backup numbers and in current implementation is 3, the Chord ring will automatically get fixed and continue to serve further k-v queries! You try to kill some nodes to see the change on visualizer.

Enjoy Chording!
