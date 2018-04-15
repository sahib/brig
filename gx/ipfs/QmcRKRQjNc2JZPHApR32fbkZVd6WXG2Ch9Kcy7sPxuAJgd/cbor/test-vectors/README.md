test-vectors
============

This repo collects some simple test vectors in machine-processable form.

appendix_a.json
---------------

All examples in Appendix A of RFC 7049, encoded as a JSON array.

Each element of the test vector is a map (JSON object) with the keys:

- cbor: a base-64 encoded CBOR data item
- hex: the same CBOR data item in hex encoding
- roundtrip: a boolean that indicates whether a generic CBOR encoder
  would _typically_ produce identical CBOR on re-encoding the decoded
  data item (your mileage may vary)
- decoded: the decoded data item if it can be represented in JSON
- diagnostic: the representation of the data item in CBOR diagnostic notation, otherwise

To make use of the cases that need diagnostic notation, a diagnostic
notation printer is usually all that is needed: decode the CBOR, print
the decoded data item in diagnostic notation, and compare.

(Note that the diagnostic notation uses full decoration for the
indefinite length byte string, while the decoded indefinite length
text string represented in JSON necessarily doesn't.)
