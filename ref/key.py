from secp256k1 import PrivateKey, PublicKey
import binascii
import sha3
N_ADDRESS_BYTES = 20
N_PUB_KEY_BYTES = 64
address = 0x6e5ab887860e199b91b92d81f418c95d9ffd32cb
key = "40dad29726f7e1b56359d2f1cc5a5365cb105b410e1108b3da65c1d97bfe6f8e"
# b = long_to_bytes(key)
privkey = PrivateKey(bytes(bytearray.fromhex(key)), raw=True)
privkey_der = privkey.serialize()
assert privkey.deserialize(privkey_der) == privkey.private_key
print "privkey",privkey.serialize()
pubkey = privkey.pubkey
pub = pubkey.serialize(compressed=False)
print "pubkey",binascii.hexlify(pub[1:])
print "pubkey",sha3.keccak_256(pub[1:]).hexdigest()[-2 * N_ADDRESS_BYTES:]
