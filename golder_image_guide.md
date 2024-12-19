## Wow wow wow this is the step by step guide so you never forget how to make a golden image from mender raspberrypi

If you are not lazy, original guide is [here](https://docs.mender.io/operating-system-updates-debian-family/convert-a-mender-debian-image). For the rest of us, here is the step by step guide.

sooo download the mender-compose thingy

```bash
git clone -b 4.2.3 https://github.com/mendersoftware/mender-convert.git
```

setup the environment var for mender compose location

```bash
cd mender-convert
MENDER_CONVERT_LOCATION=${PWD}
```

mkdir -p input

choose hosting region, from either 'eu' and 'us' ('us' if nothing is specified)
```bash
$MENDER_CONVERT_LOCATION/scripts/bootstrap-rootfs-overlay-production-server.sh \
    --output-dir $MENDER_CONVERT_LOCATION/input/rootfs_overlay_production \
    --server-url https://mender.matta.ai
```

Move your golden disk image into an input subdirectory:
```bash
mv <PATH_TO_MY_GOLDEN_IMAGE> input/golden-image-1.img
```

Build the docker before using the mender compose

```bash
./docker-build
```

Run mender-convert from inside the container with your desired options, e.g.

```bash
# move overlay to the input folder
mkdir -p input/overlay
mv <PATH_TO_MY_OVERLAY>/* input/overlay/*

# convert the image
MENDER_ARTIFACT_NAME=release-1 ./docker-mender-convert \
    --disk-image input/golden-image-1.img \
    --config configs/raspberrypi3_config \
    --overlay input/rootfs_overlay_demo/
```

Now this will generate a number of files, please proceed to making new realese using .mender file on the prod mender server.
