import argparse
import torch
import torch.nn.functional as F
from skimage import io, transform
from torchvision import datasets, transforms
from mnist import Net
from pathlib import Path
from PIL import Image


def save_image(data_loader, device, filepath):
    filepath = Path(str(filepath))
    n_image = 100
    count = 0
    for batch_data, batch_target in data_loader:
        batch_data = batch_data.to(device)
        for i in range(1000):
            print('count = {}'.format(count))
            data = batch_data.numpy()[i, :, :, :].squeeze() * 256
            Image.fromarray(data).convert('L').save(filepath / (str(count) + '.png'))
            # io.imsave(str(filepath / str(count)), batch_data.numpy()[i, :, :, :].squeeze())

            count += 1
            if count == n_image:
                return



def main():
    # Training settings
    parser = argparse.ArgumentParser(description='PyTorch MNIST Example')
    parser.add_argument('--batch-size', type=int, default=64, metavar='N',
                        help='input batch size for training (default: 64)')
    parser.add_argument('--test-batch-size', type=int, default=1000, metavar='N',
                        help='input batch size for testing (default: 1000)')
    parser.add_argument('--epochs', type=int, default=2, metavar='N',
                        help='number of epochs to train (default: 10)')
    parser.add_argument('--lr', type=float, default=0.01, metavar='LR',
                        help='learning rate (default: 0.01)')
    parser.add_argument('--momentum', type=float, default=0.5, metavar='M',
                        help='SGD momentum (default: 0.5)')
    parser.add_argument('--no-cuda', action='store_true', default=False,
                        help='disables CUDA training')
    parser.add_argument('--seed', type=int, default=1, metavar='S',
                        help='random seed (default: 1)')
    parser.add_argument('--log-interval', type=int, default=10, metavar='N',
                        help='how many batches to wait before logging training status')

    parser.add_argument('--save-model', action='store_true', default=True,
                        help='For Saving the current Model')
    args = parser.parse_args()

    torch.manual_seed(args.seed)

    use_cuda = not args.no_cuda and torch.cuda.is_available()
    device = torch.device("cuda" if use_cuda else "cpu")
    kwargs = {'num_workers': 1, 'pin_memory': True} if use_cuda else {}

    tf = transforms.Compose([
        transforms.ToTensor(),
        # transforms.Normalize((0.1307,), (0.3081,))
    ])
    test_loader = torch.utils.data.DataLoader(
        datasets.MNIST('../data', train=False, transform=tf),
        batch_size=args.test_batch_size, shuffle=True, **kwargs)

    model = Net().to(device)
    model.load_state_dict(torch.load("mnist_cnn.pt"))

    save_image(test_loader, device, "images")


if __name__ == '__main__':
    main()
