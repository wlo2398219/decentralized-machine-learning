## Prererequisites
- Install python3
- Install virtualenv
- Download hidden_layer_train_split.zip and  hidden_layer_test.csv 

## Directory & File Structure
```
finalproject
+-- rand1.py
+-- main.go and other go files
+-- mnist_training.sh
+-- client
|   +-- main.go

demo
+-- hidden_layer_test.csv
+-- mnist_training.sh
+-- hidden_layer_train_split
|   +-- hidden_layer_train_split_0.csv ~ hidden_layer_train_split_9.csv
```
## Place files
Before the training, we need to create a new directory demo/ like above and put the following directory & files
1. hidden_layer_train_split/ (to be dowloaded)
2. hidden_layer_test.csv (to be dowloaded)
3. mnist_training.sh (from finalproject)


## Execution
Run the mnist_training.sh in demo/ and you can observe the training at demo/A/A.out
- Distributed version
-- ./mnist_training.sh distributed
- Byzantine version
-- ./mnist_training.sh byzantine
