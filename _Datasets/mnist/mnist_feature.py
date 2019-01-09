import torch
import torch.nn as nn
import torch.nn.functional as F
from torchvision import transforms
from PIL import Image
import os
import sys


dir_path = os.path.dirname(os.path.realpath(__file__))


class Net(nn.Module):
    def __init__(self):
        super(Net, self).__init__()
        self.conv1 = nn.Conv2d(1, 20, 5, 1)
        self.conv2 = nn.Conv2d(20, 50, 5, 1)
        self.fc1 = nn.Linear(4 * 4 * 50, 500)
        self.fc2 = nn.Linear(500, 10)

    def forward(self, x):
        x = F.relu(self.conv1(x))
        x = F.max_pool2d(x, 2, 2)
        x = F.relu(self.conv2(x))
        x = F.max_pool2d(x, 2, 2)
        x = x.view(-1, 4 * 4 * 50)
        hidden = F.relu(self.fc1(x))
        x = self.fc2(hidden)
        return F.log_softmax(x, dim=1), hidden


def extract(file):
    device = torch.device('cpu')
    model = Net().to(device)
    model.load_state_dict(torch.load(dir_path + '/mnist_cnn.pt'))
    model.eval()

    tf = transforms.Compose([
        transforms.ToTensor(),
        transforms.Normalize((0.1307,), (0.3081,))
    ])

    with torch.no_grad():
        img = Image.open(dir_path + '/' + file)
        tensor = tf(img).unsqueeze(0).to(device)
        output, hidden = model(tensor)
        feature = hidden.numpy()
        print(','.join([str(x) for x in feature.flatten()]), end='')
        # print(output)


def main():
    file = 'images/0.png'
    if len(sys.argv) > 1:
        file = sys.argv[1]
    extract(file)


if __name__ == '__main__':
    main()
