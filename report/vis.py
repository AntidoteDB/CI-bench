import argparse
import pandas as pd
import numpy as np
import seaborn as sns
import os.path
import matplotlib.pyplot as plt

from pandas.plotting import register_matplotlib_converters


def q99(a):
    return np.quantile(a, .99)


if __name__ == "__main__":
    register_matplotlib_converters()

    parser = argparse.ArgumentParser(description='Show graphs.')
    parser.add_argument('-o', dest='output', help='output folder')
    parser.add_argument('-i', dest='input', help='input path')

    args = parser.parse_args()

    input_path = args.input
    output_folder = args.output

    # load dataframe
    df = pd.read_csv(input_path, index_col='start')
    df.index = pd.to_datetime(df.index, unit='ns', dayfirst=True)
    df['failed'] = df['failed'].astype(int)
    df['start'] = df.index
    df['latency'] = df['latency'] * 1E-6

    # group 10 seconds of datapoints
    df['round'] = df.index.floor('10S')
    df['roundS'] = df.index.floor('S')

    sns.set(style="darkgrid")

    mean_fig, mean_ax = plt.subplots()
    sns.lineplot(x="round", y="latency", data=df, ci="sd", estimator="mean", ax=mean_ax)
    mean_ax.set_title('mean')
    mean_ax.set_xlabel('time')
    mean_ax.set_ylabel('latency ms')
    mean_fig.savefig(os.path.join(output_folder, "mean.png"))

    max_fig, max_ax = plt.subplots()
    sns.lineplot(x="round", y="latency", data=df, ci="sd", estimator="max", ax=max_ax)
    max_ax.set_title('max')
    max_ax.set_xlabel('time')
    max_ax.set_ylabel('latency ms')
    max_fig.savefig(os.path.join(output_folder, "max.png"))

    quantile_fig, quantile_ax = plt.subplots()
    sns.lineplot(x="round", y="latency", data=df, ci="sd", estimator=q99, ax=quantile_ax)
    quantile_ax.set_title('99% quantile latency')
    quantile_ax.set_xlabel('time')
    quantile_ax.set_ylabel('latency ms')
    quantile_fig.savefig(os.path.join(output_folder, "99.png"))

    throughput_fig, throughput_ax = plt.subplots()
    sns.lineplot(x="roundS", y="latency", data=df, ci="sd", estimator="count", ax=throughput_ax, hue="failed")
    throughput_ax.set_title('throughput')
    throughput_ax.set_xlabel('time')
    throughput_ax.set_ylabel('throughput')
    throughput_fig.savefig(os.path.join(output_folder, "throughput.png"))
