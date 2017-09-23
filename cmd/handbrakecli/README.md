Container with the HandbrakeCLI and a preset suitable for Direct Play on
the Plex app on a Tivo Bolt.

```
/usr/bin/HandBrakeCLI --preset-import-file /config/ghb/presets.json \
    -i /path/to/inputfile \
    -o /path/to/outputfile \
    --preset tivo
```