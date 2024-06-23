import os
import csv
import argparse

LOG_DIR = os.getcwd()


def parse_args():
    parser = argparse.ArgumentParser()
    parser.add_argument(
        "-comp", "-c", help="component to search for", type=str, required=True
    )
    parser.add_argument("-level", "-l", help="log level", type=str, required=True)
    parser.add_argument("-file", "-f", help="Specific log file to parse", type=str)
    args = parser.parse_args()
    return args


# parse a single log file
def parse_log(component: str, level: str, file: str) -> None:
    results = ""
    with open(file=file, mode="r") as f:
        reader = csv.reader(f)
        for row in reader:
            if row[0] == component and row[1] == level:
                results += f"{component} - {level}: {row[3]}\n"
    print(results)


# search logs for a component with a specified log level
def parse_logs(component: str, level: str, file: str) -> None:
    results = ""
    log_files = os.listdir(LOG_DIR)

    # parse a single file
    if file:
        return parse_log(component, level, file)

    # parse all the files
    for log_file in log_files:
        if log_file.endswith(".py"):
            continue
        with open(file=log_file, mode="r") as f:
            reader = csv.reader(f)
            for row in reader:
                if row[0] == component and row[1] == level:
                    results += f"{component} - {level}: {row[3]}\n"

    print(results)


if __name__ == "__main__":
    args = parse_args()
    parse_logs(args.comp, args.level, args.file)
