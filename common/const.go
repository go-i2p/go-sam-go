package common

// DEFAULT_SAM_MIN specifies the minimum supported SAM protocol version.
// This constant is used during SAM bridge handshake to negotiate protocol compatibility.
const (
	DEFAULT_SAM_MIN = "3.1"
	// DEFAULT_SAM_MAX specifies the maximum supported SAM protocol version.
	// This allows the library to work with newer SAM protocol features when available.
	DEFAULT_SAM_MAX = "3.3"
)

// SESSION_OK indicates successful session creation with destination key.
// SESSION_DUPLICATE_ID indicates session creation failed due to duplicate session ID.
// SESSION_DUPLICATE_DEST indicates session creation failed due to duplicate destination.
// SESSION_INVALID_KEY indicates session creation failed due to invalid destination key.
// SESSION_I2P_ERROR indicates session creation failed due to I2P router error.
const (
	SESSION_OK             = "SESSION STATUS RESULT=OK DESTINATION="
	SESSION_DUPLICATE_ID   = "SESSION STATUS RESULT=DUPLICATED_ID\n"
	SESSION_DUPLICATE_DEST = "SESSION STATUS RESULT=DUPLICATED_DEST\n"
	SESSION_INVALID_KEY    = "SESSION STATUS RESULT=INVALID_KEY\n"
	SESSION_I2P_ERROR      = "SESSION STATUS RESULT=I2P_ERROR MESSAGE="
)

// Signature Type Constants - I2P Cryptographic Security Configuration
//
// SECURITY RECOMMENDATION: Always use SIG_DEFAULT (EdDSA_SHA512_Ed25519) for new applications.
// EdDSA provides superior performance, smaller key sizes, and robust security compared to
// legacy signature algorithms. It is the I2P network's recommended signature type.
//
// SIG_NONE is deprecated, use SIG_DEFAULT instead for secure signatures.
// SIG_DSA_SHA1 specifies DSA with SHA1 signature type (LEGACY - NOT RECOMMENDED for new applications).
//   - Legacy algorithm with known cryptographic weaknesses
//   - Larger key sizes and slower performance
//   - Should only be used for compatibility with very old I2P destinations
//
// SIG_ECDSA_SHA256_P256 specifies ECDSA with SHA256 on P256 curve signature type.
//   - Acceptable security but larger signatures than EdDSA
//   - Consider EdDSA for better performance
//
// SIG_ECDSA_SHA384_P384 specifies ECDSA with SHA384 on P384 curve signature type.
//   - Higher security margin but significantly larger signatures
//   - Slower key generation and verification
//
// SIG_ECDSA_SHA512_P521 specifies ECDSA with SHA512 on P521 curve signature type.
//   - Highest security but largest signatures and slowest performance
//   - Only recommended for extremely high-security applications
//
// SIG_EdDSA_SHA512_Ed25519 specifies EdDSA with SHA512 on Ed25519 curve signature type.
//   - RECOMMENDED: Fastest signature verification, smallest signatures
//   - State-of-the-art cryptographic security with excellent performance
//   - Default choice for all new I2P applications
//
// SIG_DEFAULT points to the recommended secure signature type for new applications.
//   - Currently set to EdDSA_SHA512_Ed25519 for optimal security and performance
const (
	SIG_NONE                 = "SIGNATURE_TYPE=EdDSA_SHA512_Ed25519"
	SIG_DSA_SHA1             = "SIGNATURE_TYPE=DSA_SHA1"
	SIG_ECDSA_SHA256_P256    = "SIGNATURE_TYPE=ECDSA_SHA256_P256"
	SIG_ECDSA_SHA384_P384    = "SIGNATURE_TYPE=ECDSA_SHA384_P384"
	SIG_ECDSA_SHA512_P521    = "SIGNATURE_TYPE=ECDSA_SHA512_P521"
	SIG_EdDSA_SHA512_Ed25519 = "SIGNATURE_TYPE=EdDSA_SHA512_Ed25519"
	// Add a default constant that points to the recommended secure signature type
	SIG_DEFAULT = SIG_EdDSA_SHA512_Ed25519
)

// SESSION_ADD_OK indicates successful subsession addition to primary session.
// SESSION_REMOVE_OK indicates successful subsession removal from primary session.
const (
	SESSION_ADD_OK    = "SESSION STATUS RESULT=OK"
	SESSION_REMOVE_OK = "SESSION STATUS RESULT=OK"
)

// SAM_RESULT_OK indicates successful SAM operation completion.
// SAM_RESULT_INVALID_KEY indicates SAM operation failed due to invalid key format.
// SAM_RESULT_KEY_NOT_FOUND indicates SAM operation failed due to missing key.
const (
	SAM_RESULT_OK            = "RESULT=OK"
	SAM_RESULT_INVALID_KEY   = "RESULT=INVALID_KEY"
	SAM_RESULT_KEY_NOT_FOUND = "RESULT=KEY_NOT_FOUND"
)

// HELLO_REPLY_OK indicates successful SAM handshake completion.
// HELLO_REPLY_NOVERSION indicates SAM handshake failed due to unsupported protocol version.
const (
	HELLO_REPLY_OK        = "HELLO REPLY RESULT=OK"
	HELLO_REPLY_NOVERSION = "HELLO REPLY RESULT=NOVERSION\n"
)

// SESSION_STYLE_STREAM creates TCP-like reliable connection sessions.
// SESSION_STYLE_DATAGRAM creates UDP-like message-based sessions.
// SESSION_STYLE_RAW creates low-level packet transmission sessions.
const (
	SESSION_STYLE_STREAM   = "STREAM"
	SESSION_STYLE_DATAGRAM = "DATAGRAM"
	SESSION_STYLE_RAW      = "RAW"
)

// ACCESS_TYPE_WHITELIST allows only specified destinations in access list.
// ACCESS_TYPE_BLACKLIST blocks specified destinations in access list.
// ACCESS_TYPE_NONE disables access list filtering entirely.
const (
	ACCESS_TYPE_WHITELIST = "whitelist"
	ACCESS_TYPE_BLACKLIST = "blacklist"
	ACCESS_TYPE_NONE      = "none"
)
