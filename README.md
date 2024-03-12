
<p align="center">
<b align="center" style="font-size: 4em;"> 👝 </b><br />
<b>fpouch</b><br /> Minimalist file sharing and upload server.
</p>

(Alpha stage) 

### How to use

Simple file server:

```
./fpouch --store-path ./media
```

Without `--store-path` it will store file in current working directory. If the supplied, but the folder doesn't exist it will created.

Setup only for upload option:

```
./fpouch --no-sharing
```

Run only for sharing option:

```
./fpouch --no-upload
```

No UI option, you may direct upload from endpoint (e.g using CURL) and sharing index in JSON format.

```
./fpouch --no-ui
```

### License

[UNLICENSE](./UNLICENSE)