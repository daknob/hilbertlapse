# hilbertlapse
hilbertlapse is a collection of tools that are used to image IP Space, at a
fixed point in time, or as a time lapse, combining multiple snapshots into a
single animated GIF.

The pipeline used by hilbertlapse is the following:

```
+------+     +-------+     +---------+
| SCAN +---->+ IMAGE +---->+ ANIMATE |
+------+     +-------+     +---------+
```

Therefore, the entire project is not a single binary that runs and outputs the
desired results, but instead three binaries, or at least three steps, which are
connected either by a different program, or simple bash scripts and pipes.

This is done to add flexibility to the entire process, so any binary in the
pipeline can be exchanged with any other, interchangeably, but also to make it
easier to develop and prototype, as this was created for a [blog
post](https://blog.daknob.net/mapping-the-greek-inet-oct-19/). In that blog
post you can find more information about the project. This flexibility now
allows hilbertcurve to run each part in a different machine, or set of
machines. For example, run pings in a set of servers, and imaging and animating
in a researcher laptop.

The three steps of hilbertlapse are described below:

## Scanner

The scanner is a utility that performs some sort of scan on a set of IP
Addresses. The current recommended way to achieve this is using
[masscan](https://github.com/robertdavidgraham/masscan), which works on many
operating systems, is one of the fastest pieces of software, and it's available
in many repositories.

The recommended way to run `masscan` is:

```
/usr/bin/masscan --ping --rate=256 -oL "scans/$(date +%Y/%m/%d/%H:%M)" 193.5.16.0/22 2>/dev/null > /dev/null
```

You can of course adjust the rate, do sharding, etc. as long as you use the
list (`-oL`) output of the tool and you have the results available in a single
file.  This would also work with port scans, but ping is used above as an
example.

The above uses the `date` command to name the output file so it's safe to run
in a `cron` job e.g. every hour, producing a unique file per scan. In this
example, you'd have a `scans` directory, and within it one folder per year,
containing one folder per month, containing one folder per day, containing all
the hourly scans. Be sure to `mkdir -p` the folder first!

An example output file would be:

```
open icmp 0 193.5.16.0 1703671357
open icmp 0 193.5.16.1 1703671358
open icmp 0 193.5.16.80 1703671358
open icmp 0 193.5.16.89 1703671358
```

## Imager

The `imager` is a utility that accepts a file from masscan, and then creates a
PNG, based on the details included in the input. The `imager` included in the
current repository will take a scan of a network and create a PNG that will
have each `up` value set to `#32c832` and each `down` to the color `#323232`.
The colors are of course configurable.

The mapping of IPv4 Addresses to `(X,Y)` coordinates on the PNG file is done
using a [Hilbert Curve](https://en.wikipedia.org/wiki/Hilbert_curve) map, to
add subnet locality.

An example output of `imager` is this:

![Example imager
output](https://blog.daknob.net/content/images/2019/11/04/uoc.png)

## Animator
The final component of the pipeline, the `animator`, accepts a list of file
names of PNG files, and then creates and outputs an Animated GIF, where each
frame is the content of every PNG that has been given to it. The frames are
added in the order the file name appeared on the input, so you can change the
order of frames by controlling the order of the PNG file names to the program
input.

Each frame in the GIF has a delay of `0`, so the frames can go by quickly, and
make longer scans or very small scan intervals be relatively short GIFs and
relatively small files.

An example output of `animator` is this:

![Animator example
output](https://blog.daknob.net/content/images/2020/01/03/36C3.gif)

# Examples
You can find more information and more real world examples in my [main blog
post](https://blog.daknob.net/mapping-the-greek-inet-oct-19/), as well as a
subsequent one about [the scan of the #36C3
event](https://blog.daknob.net/mapping-36c3/).

# How to run
After you build and compile the three tools above (or yours), you can have a
running setup that will be doing the scans fairly easily. The three tools can
run on the same computer, or on different ones, it does not matter, and the
scanning can be parallelized over multiple hosts.

## Scanning
First, in the scanning computer, add a cron job, or any other means of
preference to execute a script periodically. This could run every period you
want. I have used one hour for mine. The script that should run should be:

```bash
pinger -a 147 -b 52 -g 128
```

The variables `a` and `b` are the /8 and the /16 part of the network
respectively (the first and second octet of the IPv4 address). The `g` flag is
the amount of parallel scans you want to run. If you have this number too
small, the scan may not complete in time, if you have this number too large,
you may trigger firewalls, filters, etc. and have your address blocked, or your
computer may not be able to handle the output and drop packets, giving you
false results. The default value, `128`, takes a few minutes to scan a `/16`
that is < 100 ms away, with relative ease, and the results do not differ from a
non-parallel scan, for all target networks tried.

This will generate a lot of files, one for every scan. If you run this with a
very small interval or for a very long time, be careful as you may cause
problems to your filesystem, and reach its limits. It is safe to move results
to different folders while the script runs, after it is done, NOT while it
runs. You can therefore write another cron job that will move files to multiple
folders, based, for example, on their day of the month.

## Imaging
The next phase is imaging, here you need to run the `imager` once for every
scan you want to create a PNG of. The `imager` accepts the file name, of the
input CSV, and the file name of the output PNG.

An example of doing this in `bash` can be:

```bash
for f in $(ls -1|grep csv); do imager -i $f -o pngs/$f.png; done
```

This will iterate over all `csv` files (or files whose name contains the word
`csv`, feel free to make the `grep` more strict or loose depending on your
needs), and call `imager` for each one, giving it the input file, and having
the output be written in a directory called `pngs`, with the file name being
the CSV file name, with the `.png` suffix to the end.

Feel free to parallelize as you see fit, for example by doing `.png & done` or
using `xargs`. This runs pretty fast anyways, so there has been no need to do
that so far.

## Animating
The final step is to, if required, create an animated map of the PNGs, in the
form of a GIF. In order to do this, you can use the `animator`, which accepts
a **list of PNG file names** on its input, and then creates an animated GIF
with the order the file names came in.

You can call the `animator` in the following way:

```bash
ls -1 | grep png | sort | animator -o uofcrete.gif
```

You need to run the command above on the `pngs` folder created above, or adjust
the `ls` command appropriately. You can `grep` for anything you want, such as
specific dates, or anything you can imagine. You can `sort` for them to be in
chronological order, or `sort -r` to be in reverse order. You can generate the
file name list any way you want, and do any processing you require, either with
a program, or standard UNIX utilities.

After that, you send this list to the `animator`, which will take it, and given
the output file name, here `uofcrete.gif`, you will receive the final animated
GIF at the file you require.

# Contributing
Feel free to contribute any changes you may have made, or open GitHub Issues
with any requests, problems, bugs, etc. If you write additional scanners,
imagers, or animators, they are more than welcome and I will appreciate any
Pull Requests. Make sure they are in their own folder, and do not have any
dependencies you did not include. Ideally make them CSV-compatible with the
current ones, so people can decide which imager and scanner they want to use.
If you need to include a `README.md` file with instructions, please do so in
the folder of the tool you contribute.

Also, make sure you have seen the license you contribute your code under, and
if you are not okay to contribute under this license, please create a different
repository, containing just your tool, and link to this one here. Also, feel
free to modify this document with a link to your tool, so people can find it
and use it.

# Alternative Components
Here you can find a list of alternative Scanners, Imagers, and Animators, that,
for some reason, have their own repository. Feel free to add yours here:

## Scanner
* No known projects

## Imager
* No known projects

## Animator
* No known projects
