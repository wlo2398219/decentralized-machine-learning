## Prererequisites
- Install python3
- Install virtualenv
- Download [hidden_layer_train_uni_split.zip](https://drive.google.com/file/d/1oTq_px8un_yL4BYwsNcJqofRpIYFtV3w/view?fbclid=IwAR1LnWR7-cQ_SE0BnR462n2J-aYml00GFNHmaYB99jbLZ9pNCpEYf0AXiIE) and  [hidden_layer_test.csv](https://drive.google.com/file/d/1wvjx4Vo_n37WjRdSGoAgQX4_vXHhc5KS/view?usp=sharing)
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
1. hidden_layer_train_uni_split/ (to be dowloaded)
2. hidden_layer_test.csv (to be dowloaded)
3. mnist_training.sh (from finalproject)


## Execution
Run the mnist_training.sh in demo/ and you can observe the training at demo/A/A.out
- Distributed version
-- ./mnist_training.sh distributed
- Byzantine version
-- ./mnist_training.sh byzantine

