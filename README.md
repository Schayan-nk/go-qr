# go-qr

A small, dependency-free QR code encoder for one job: putting an
authenticator app's setup address (`otpauth://...`) on a screen so a phone can
read it, without sending the secret anywhere.

It is a few hundred lines of published arithmetic (ISO/IEC 18004). It exists
so that a page showing a two-factor secret needs neither a third-party
dependency to audit nor a script from a CDN that the page's Content Security
Policy would have to allow.

- **Deliberately narrow:** byte mode, error-correction level M, versions 1 to
  10 (up to 213 bytes). An `otpauth://` address is about a hundred bytes.
  Anything longer is refused with `ErrTooLong` rather than drawn wrong.
- **Checkable:** the construction follows Project Nayuki's reference
  implementation step for step (the same tables, placement order and penalty
  rules), so it can be compared line by line.
- **Verified against an independent encoder:** the tests compare every module
  of 35 symbols (versions 1 and 5 to 10, every mask choice) with
  [segno](https://github.com/heuer/segno).

## Install

```sh
go get github.com/Schayan-nk/go-qr
```

## Example

```go
code, err := qr.Encode("otpauth://totp/Example:alice?secret=JBSWY3DPEHPK3PXP&issuer=Example")
if err != nil {
	return err
}

svg := code.SVG() // a self-contained <svg>, quiet zone included

// or draw it yourself, one module at a time:
for y := 0; y < code.Size; y++ {
	for x := 0; x < code.Size; x++ {
		if code.Dark(x, y) {
			// paint module (x, y)
		}
	}
}
```

## Tests

```sh
go test ./...
```

`testdata/segno.json` holds the reference grids. To regenerate them, install
segno and run `python testdata/segno_gen.py <segno-path> testdata/segno.json`.

## License

MIT, see [LICENSE](LICENSE).
