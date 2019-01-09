#!/bin/bash

mnistDir=$(dirname "$0")
envPath=$mnistDir"/feature-env"
relMnistPath=$1

# Create shell script to run feature extractor.
echo "#!/bin/bash
mnistDir=$relMnistPath
dataFile=\$1
source \$mnistDir/feature-env/bin/activate
python \$mnistDir/mnist_feature.py \$dataFile" > mnist_feature_extractor.sh

# Create virtual environment if needed.
echo "Look for virtual environment at "$envPath
if [ -d $envPath ]; then
    echo "Directory alread exists."
else
    # If the environment doesn't exist.
    echo "Create virtual environment for the Python feature extrator."
    virtualenv --python=python3 $envPath
    source $envPath/bin/activate
    pip3 install torch torchvision
fi
