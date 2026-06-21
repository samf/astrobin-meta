# astrobin-csv

A small Go program that makes uploads to astrobin.com a bit easier. It's highly
specific to the way that I work, and is not meant to be a general tool for
everyone. See https://github.com/SteveGreaves/AstroBinUploader for a more
complete solution.

It scans a directory of LIGHT frames (FITS or XISF) captured by N.I.N.A.,
groups them by filter, and writes a CSV in the format AstroBin's
["import acquisitions from CSV"](https://welcome.astrobin.com/importing-acquisitions-from-csv)
dialogue expects. One row is written per filter, aggregating every light frame
found for that filter across all nights/sessions under the given directory.

## Build

```sh
go build -o astrobin-csv .
```

## Configuration

Filter names (the `FILTER` keyword in each frame's header) are mapped to numeric
AstroBin filter IDs via a YAML config file at `~/.astrobin-csv.yaml`. See
[`astrobin-csv.example.yaml`](astrobin-csv.example.yaml) for the format:

```yaml
filters:
  L: 33995
  H: 43627
  O: 43628
```

Filters present in your frames but absent from the config are skipped, with a
warning listing the names to add.

## Usage

```sh
# Scan a directory recursively and write acquisition.csv
astrobin-csv /path/to/target/lights

# Custom output path
astrobin-csv /path/to/target/lights -o acquisition.csv

# Custom config path
astrobin-csv /path/to/target/lights -c /path/to/filters.yaml
```

Run `astrobin-csv --help` for the full flag list.
