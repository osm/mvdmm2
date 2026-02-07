# mvdmm2

Parse deaths, tooks and mm2 reports from an MVD demo.

## Usage

```sh
$ mvdmm2 testdata/20241210-2125_4on4_sk_vs_tot\[dm3\].mvd
name: ToT_Javve        death:  39/54  took:  16/43
name: ToT_fix          death:  38/45  took:  12/49
name: ToT_phren        death:  41/57  took:  17/41
name: ToT_slime        death:  12/15  took:  45/51
name: gosciu           death:   3/42  took:   0/50
name: kane             death:   0/63  took:   1/41
name: rokky            death:  37/54  took:   0/37
name: snapcase         death:   2/57  took:   1/34
```

```sh
$ mvdmm2 testdata/20241210-2125_4on4_sk_vs_tot\[dm3\].mvd all | head
00:00: ToT_Javve lost
00:00: kane lost
00:00: rokky took rl
00:00: gosciu took ya
00:02: ToT_phren took mega
00:02: ToT_slime took ring
00:02: snapcase took pent
00:03: ToT_fix took ng
00:04: ToT_slime took quad
00:04: ToT_slime took-report
```
