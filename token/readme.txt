Bcrypt algo
plain txt+ cost +salt= hashed password
produces different hashes each time..

Pass for eg:
algo$cost$cost$salt$hash
2l$2digits$22chars(16bits)$24bytes(31chars)
########
Security/ JWT/Paseto

JWT --Json Web Token --has some security issues

It is base 64 encrypted

having three components:
1) Signing Algorithm and tokentype
2)Payload
3)Digital Signature (this only server have private key to sign the token)
so if hacker attempts to send request with wrong secret key 
it will be detected by the server.

-->JWT signing algo's
1)Symmetric digital signature Algorithm
-->same key is used to sign and verify token.
-->for local use:internal services,where the key can be shared, eg: HS256,HS386,HS512
 HS256=HMAC+SHA256
HMAC = Hash based MEssage authentication code
SHA  = Secure Hash Algorithm
256/384/512: number of output bits
2)Asymmetric  digital secret Algorithm
-->The private key is used to sign the token
-->The public key is used to verify the token
-->For Public Use:internal service signs the token , 
but external service needs to verify it.
RS256,RS384,RS512 || PS256,PS384,PS512 || ES256,ES384,ES512

RS256 = RSA PKCSv1.5 + SHA256 [PKCS:Public-Key Cryptography Standards]
PS256 = RSA PSS + SHA256 [Probabilistic Signature Signature Scheme]
ES256 = ECDSA+ SHA256 [ECDSA:Elliptic Curve Digital Signature Algorithm]

Problem:too many algo to choose
1)RSA   : Padding Oracle attack
2)ECDSA : Invalid curve attack

Trivial forgery
-->set alg header as   "none"
-->set alg header to "HS256" while the server normally verifies token 
   with RSA public key,

   Therefore servre must check the header first before signing in the tokens

PASETO: Platform-Agnostic Security Tokens
Stronger Algorithms
Only need to select version to use por PASETO
Each version has 1 strong cipher suite
Only 2 most recent PASTO versions are accepted

v1) compatible with legacy systems:
 --> local : <symmteric key>
    Authenticated Encryption
    AES256 CTR + HMAC SHA384
 --> public: <asymmetric key>
     Digital Signature
     RSAPSS + SHA384  

-->Non Trivial Forgery:
    No more algo header or none algo
Everything is Authenticated
Encrypted payload for local use<symmetric key> 

v2:
 local:<symmetric>
    Authenticated Encryption
    XChaCha20-Poly1305
 public:<asymmetric key>  --in this payload is not encrypted but only base64 encoded
    Digital Signature
    Ed25519[EdDSA + Curve25519]

 Paseto request structure:
 Local-->
 1) Version:v2
 2) Purpose :local[symmetric-key authenticated encryption]
 3) Payload : [hex-encoded]
     Body:
        Encrypted:
        Decrypted:
     Nonce: Authorises and used for encrypting data
     Authorisatio tag: to authorise the encrypted data
4)Footer:it is encoded , for addtional info
Public:
1) Version:v2
 2) Purpose :public[asymmetric-key digital signature]
 3) Payload : 
     Body:
        Encoded[base-64]:
        Decoded:
     Signature:[hex encoded]
     
######
########################################
type SigningMethod interface {
	Verify(signingString, signature string, key interface{}) error // Returns nil if signature is valid
	Sign(signingString string, key interface{}) (string, error)    // Returns encoded signature or error
	Alg() string                                                   // returns the alg identifier for this method (example: 'HS256')
}
########################################
func NewWithClaims(method SigningMethod, claims Claims) *Token {
	return &Token{
		Header: map[string]interface{}{
			"typ": "JWT",
			"alg": method.Alg(),
		},
		Claims: claims,
		Method: method,
	}
}
############################################
func (t *Token) SignedString(key interface{}) (string, error) {
	var sig, sstr string
	var err error
	if sstr, err = t.SigningString(); err != nil {
		return "", err
	}
	if sig, err = t.Method.Sign(sstr, key); err != nil {
		return "", err
	}
	return strings.Join([]string{sstr, sig}, "."), nil
}
########################################################
The below function call takes in the token object
and encode the header and claims segment of the token and return the '.' seperated segments in string format
########################################################
func (t *Token) SigningString() (string, error) {
	var err error
	parts := make([]string, 2)
	for i, _ := range parts {
		var jsonValue []byte
		if i == 0 {
			if jsonValue, err = json.Marshal(t.Header); err != nil {
				return "", err
			}
		} else {
			if jsonValue, err = json.Marshal(t.Claims); err != nil {
				return "", err
			}
		}

		parts[i] = EncodeSegment(jsonValue)
	}
	return strings.Join(parts, "."), nil
}

// func (t *Token) SignedString(key interface{}) (string, error)
// SignedString creates and returns a complete, signed JWT.
// The token is signed using the SigningMethod specified in the token.

##########################################################################
type Claims interface {
	GetExpirationTime() (*NumericDate, error)
	GetIssuedAt() (*NumericDate, error)
	GetNotBefore() (*NumericDate, error)
	GetIssuer() (string, error)
	GetSubject() (string, error)
	GetAudience() (ClaimStrings, error)
}
type ClaimsValidator interface {
	Claims
	Validate() error
}

ClaimsValidator is an interface that can be implemented by custom claims 
who wish to execute any additional claims validation based on 
application-specific logic. The Validate function is then executed 
in addition to the regular claims validation and any error returned is 
appended to the final validation result.
################################################
type Keyfunc func(*Token) (interface{}, error)
Keyfunc will be used by the Parse methods as a callback function to supply the key
for verification. The function receives the parsed, but unverified Token. 
This allows you to use properties in the Header of the token (such as `kid`) to 
identify which key to use.
The returned interface{} may be a single key or a VerificationKeySet containing
multiple keys.

#######################
type Parser struct {
	ValidMethods         []string // If populated, only these methods will be considered valid
	UseJSONNumber        bool     // Use JSON Number format in JSON decoder
	SkipClaimsValidation bool     // Skip claims validation during token parsing
}
########################

#################################################
Parse With claims method
#################################################

func (p *Parser) ParseWithClaims(tokenString string, claims Claims, keyFunc Keyfunc) (*Token, error) {
	token, parts, err := p.ParseUnverified(tokenString, claims)
	if err != nil {
		return token, err
	}

	// Verify signing method is in the required set
	if p.ValidMethods != nil {
		var signingMethodValid = false
		var alg = token.Method.Alg()
		for _, m := range p.ValidMethods {
			if m == alg {
				signingMethodValid = true
				break
			}
		}
		if !signingMethodValid {
			// signing method is not in the listed set
			return token, NewValidationError(fmt.Sprintf("signing method %v is invalid", alg), ValidationErrorSignatureInvalid)
		}
	}

	// Lookup key
	var key interface{}
	if keyFunc == nil {
		// keyFunc was not provided.  short circuiting validation
		return token, NewValidationError("no Keyfunc was provided.", ValidationErrorUnverifiable)
	}
	if key, err = keyFunc(token); err != nil {
		// keyFunc returned an error
		if ve, ok := err.(*ValidationError); ok {
			return token, ve
		}
		return token, &ValidationError{Inner: err, Errors: ValidationErrorUnverifiable}
	}

	vErr := &ValidationError{}
if !p.SkipClaimsValidation {
		if err := token.Claims.Valid(); err != nil {

			// If the Claims Valid returned an error, check if it is a validation error,
			// If it was another error type, create a ValidationError with a generic ClaimsInvalid flag set
			if e, ok := err.(*ValidationError); !ok {
				vErr = &ValidationError{Inner: err, Errors: ValidationErrorClaimsInvalid}
			} else {
				vErr = e
			}
		}
	}

	// Perform validation
	token.Signature = parts[2]
	if err = token.Method.Verify(strings.Join(parts[0:2], "."), token.Signature, key); err != nil {
		vErr.Inner = err
		vErr.Errors |= ValidationErrorSignatureInvalid
	}

	if vErr.valid() {
		token.Valid = true
		return token, nil
	}

	return token, vErr
}
###################################################################################
ParseUnverified  Method
###################################################################################
func (p *Parser) ParseUnverified(tokenString string, claims Claims) (token *Token, parts []string, err error) {
	parts = strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, parts, NewValidationError("token contains an invalid number of segments", ValidationErrorMalformed)
	}

	token = &Token{Raw: tokenString}

	// parse Header
	var headerBytes []byte
	if headerBytes, err = DecodeSegment(parts[0]); err != nil {
		if strings.HasPrefix(strings.ToLower(tokenString), "bearer ") {
			return token, parts, NewValidationError("tokenstring should not contain 'bearer '", ValidationErrorMalformed)
		}
		return token, parts, &ValidationError{Inner: err, Errors: ValidationErrorMalformed}
	}
	if err = json.Unmarshal(headerBytes, &token.Header); err != nil {
		return token, parts, &ValidationError{Inner: err, Errors: ValidationErrorMalformed}
	}

	// parse Claims
	var claimBytes []byte
	token.Claims = claims

	if claimBytes, err = DecodeSegment(parts[1]); err != nil {
		return token, parts, &ValidationError{Inner: err, Errors: ValidationErrorMalformed}
	}
	dec := json.NewDecoder(bytes.NewBuffer(claimBytes))
	if p.UseJSONNumber {
		dec.UseNumber()
	}
	// JSON Decode.  Special case for map type to avoid weird pointer behavior
	if c, ok := token.Claims.(MapClaims); ok {
		err = dec.Decode(&c)
	} else {
		err = dec.Decode(&claims)
	}

    if err != nil {
		return token, parts, &ValidationError{Inner: err, Errors: ValidationErrorMalformed}
	}

	// Lookup signature method
	if method, ok := token.Header["alg"].(string); ok {
		if token.Method = GetSigningMethod(method); token.Method == nil {
			return token, parts, NewValidationError("signing method (alg) is unavailable.", ValidationErrorUnverifiable)
		}
	} else {
		return token, parts, NewValidationError("signing method (alg) is unspecified.", ValidationErrorUnverifiable)
	}

	return token, parts, nil
}


###########

Flow-->CreateToken func

       NewWithClaims
       Signed String
       func (m *SigningMethodHMAC) Sign(signingString string, key interface{}) (string, error) {
	if keyBytes, ok := key.([]byte); ok {
		if !m.Hash.Available() {
			return "", ErrHashUnavailable
		}

		hasher := hmac.New(m.Hash.New, keyBytes)////
		hasher.Write([]byte(signingString))

		return EncodeSegment(hasher.Sum(nil)), nil
	}

	return "", ErrInvalidKeyType
}
###############
type SigningMethodHMAC struct {
	Name string
	Hash crypto.Hash
}
#########################
hmac.New()
func New(h func() hash.Hash, key []byte) hash.Hash {
	if boring.Enabled {
		hm := boring.NewHMAC(h, key)
		if hm != nil {
			return hm
		}
		// BoringCrypto did not recognize h, so fall through to standard Go code.
	}
	hm := new(hmac)
	hm.outer = h()
	hm.inner = h()
	unique := true
	func() {
		defer func() {
			// The comparison might panic if the underlying types are not comparable.
			_ = recover()
		}()
		if hm.outer == hm.inner {
			unique = false
		}
	}()
	if !unique {
		panic("crypto/hmac: hash generation function does not produce unique values")
	}
	blocksize := hm.inner.BlockSize()
	hm.ipad = make([]byte, blocksize)
	hm.opad = make([]byte, blocksize)
	if len(key) > blocksize {
		// If key is too big, hash it.
		hm.outer.Write(key)
		key = hm.outer.Sum(nil)
	}
	copy(hm.ipad, key)
	copy(hm.opad, key)
	for i := range hm.ipad {
		hm.ipad[i] ^= 0x36
	}
	for i := range hm.opad {
		hm.opad[i] ^= 0x5c
	}
	hm.inner.Write(hm.ipad)

	return hm
}


########################################################################
package hmac

import (
	"crypto/internal/boring"
	"crypto/subtle"
	"hash"
)

// FIPS 198-1:
// https://csrc.nist.gov/publications/fips/fips198-1/FIPS-198-1_final.pdf

// key is zero padded to the block size of the hash function
// ipad = 0x36 byte repeated for key length
// opad = 0x5c byte repeated for key length
// hmac = H([key ^ opad] H([key ^ ipad] text))

// marshalable is the combination of encoding.BinaryMarshaler and
// encoding.BinaryUnmarshaler. Their method definitions are repeated here to
// avoid a dependency on the encoding package.
type marshalable interface {
	MarshalBinary() ([]byte, error)
	UnmarshalBinary([]byte) error
}

type hmac struct {
	opad, ipad   []byte
	outer, inner hash.Hash

	// If marshaled is true, then opad and ipad do not contain a padded
	// copy of the key, but rather the marshaled state of outer/inner after
	// opad/ipad has been fed into it.
	marshaled bool
}

func (h *hmac) Sum(in []byte) []byte {
	origLen := len(in)
	in = h.inner.Sum(in)

	if h.marshaled {
		if err := h.outer.(marshalable).UnmarshalBinary(h.opad); err != nil {
			panic(err)
		}
	} else {
		h.outer.Reset()
		h.outer.Write(h.opad)
	}
	h.outer.Write(in[origLen:])
	return h.outer.Sum(in[:origLen])
}

func (h *hmac) Write(p []byte) (n int, err error) {
	return h.inner.Write(p)
}

func (h *hmac) Size() int      { return h.outer.Size() }
func (h *hmac) BlockSize() int { return h.inner.BlockSize() }

func (h *hmac) Reset() {
	if h.marshaled {
		if err := h.inner.(marshalable).UnmarshalBinary(h.ipad); err != nil {
			panic(err)
		}
		return
	}

	h.inner.Reset()
	h.inner.Write(h.ipad)

	// If the underlying hash is marshalable, we can save some time by
	// saving a copy of the hash state now, and restoring it on future
	// calls to Reset and Sum instead of writing ipad/opad every time.
	//
	// If either hash is unmarshalable for whatever reason,
	// it's safe to bail out here.
	marshalableInner, innerOK := h.inner.(marshalable)
	if !innerOK {
		return
	}
	marshalableOuter, outerOK := h.outer.(marshalable)
	if !outerOK {
		return
	}

	imarshal, err := marshalableInner.MarshalBinary()
	if err != nil {
		return
	}

	h.outer.Reset()
	h.outer.Write(h.opad)
	omarshal, err := marshalableOuter.MarshalBinary()
	if err != nil {
		return
	}

	// Marshaling succeeded; save the marshaled state for later
	h.ipad = imarshal
	h.opad = omarshal
	h.marshaled = true
}

// New returns a new HMAC hash using the given hash.Hash type and key.
// New functions like sha256.New from crypto/sha256 can be used as h.
// h must return a new Hash every time it is called.
// Note that unlike other hash implementations in the standard library,
// the returned Hash does not implement encoding.BinaryMarshaler
// or encoding.BinaryUnmarshaler.
func New(h func() hash.Hash, key []byte) hash.Hash {
	if boring.Enabled {
		hm := boring.NewHMAC(h, key)
		if hm != nil {
			return hm
		}
		// BoringCrypto did not recognize h, so fall through to standard Go code.
	}
	hm := new(hmac)
	hm.outer = h()
	hm.inner = h()
	unique := true
	func() {
		defer func() {
			// The comparison might panic if the underlying types are not comparable.
			_ = recover()
		}()
		if hm.outer == hm.inner {
			unique = false
		}
	}()
	if !unique {
		panic("crypto/hmac: hash generation function does not produce unique values")
	}
	blocksize := hm.inner.BlockSize()
	hm.ipad = make([]byte, blocksize)
	hm.opad = make([]byte, blocksize)
	if len(key) > blocksize {
		// If key is too big, hash it.
		hm.outer.Write(key)
		key = hm.outer.Sum(nil)
	}
	copy(hm.ipad, key)
	copy(hm.opad, key)
	for i := range hm.ipad {
		hm.ipad[i] ^= 0x36
	}
	for i := range hm.opad {
		hm.opad[i] ^= 0x5c
	}
	hm.inner.Write(hm.ipad)

	return hm
}

// Equal compares two MACs for equality without leaking timing information.
func Equal(mac1, mac2 []byte) bool {
	// We don't have to be constant time if the lengths of the MACs are
	// different as that suggests that a completely different hash function
	// was used.
	return subtle.ConstantTimeCompare(mac1, mac2) == 1
}