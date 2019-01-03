import random
import numpy as np


def generate():
    # y = x * w + noise
    
    dim = 16
    n = 1000
    w = np.random.rand(dim, 1)
    x = np.random.rand(n, dim) * 10
    y = np.dot(x, w) + 0.01 * np.random.rand(n, 1)
    
    print(np.dot(x, w).shape)
    
    with open('./fake.txt', 'w') as f:
        for i in range(n):
            # s = '   '.join([''] + [str(v) for v in x[i, :8]] + ['0'] + [str(v) for v in x[i, 8:]] + [str(y[i, 0])])
            s = ','.join([str(v) for v in x[i]] + [str(y[i, 0])])
            f.write(s + '\n')


if __name__ == '__main__':
    generate()