package security

import (
	"bytes"

	"git.condensat.tech/bank/database/model"
	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/armor"
	"golang.org/x/crypto/openpgp/packet"
)

func CreateKeys(name, comment, email string) (model.PgpPublicKey, model.PgpPrivateKey, error) {
	key, err := openpgp.NewEntity(name, comment, email, nil)
	if err != nil {
		return "", "", err
	}
	pubKey, err := WritePublicKey(key)
	if err != nil {
		return "", "", err
	}

	privKey, err := WritePrivateKey(key)
	if err != nil {
		return "", "", err
	}

	return pubKey, privKey, nil
}

func readReadKey(key string) *openpgp.Entity {
	reader := bytes.NewBuffer([]byte(key))

	block, err := armor.Decode(reader)
	if err != nil {
		return nil
	}

	entity, err := openpgp.ReadEntity(packet.NewReader(block.Body))
	if err != nil {
		return nil
	}
	return entity
}

func ReadPrivateKey(privKey model.PgpPrivateKey) *openpgp.Entity {
	return readReadKey(string(privKey))
}

func ReadPublicKey(pubKey model.PgpPublicKey) *openpgp.Entity {
	return readReadKey(string(pubKey))
}

func WritePublicKey(entity *openpgp.Entity) (model.PgpPublicKey, error) {
	key, err := writeEntity(entity, false)
	return model.PgpPublicKey(key), err
}

func WritePrivateKey(entity *openpgp.Entity) (model.PgpPrivateKey, error) {
	key, err := writeEntity(entity, true)
	return model.PgpPrivateKey(key), err
}

func writeEntity(entity *openpgp.Entity, private bool) (string, error) {
	blockType := "PGP PUBLIC KEY BLOCK"
	if private {
		blockType = "PGP PRIVATE KEY BLOCK"
	}

	var err error
	buf := bytes.NewBuffer(nil)
	if private {
		err = entity.SerializePrivate(buf, nil)
	} else {
		err = entity.Serialize(buf)
	}
	if err != nil {
		return "", err
	}

	out := bytes.NewBuffer(nil)
	arm, err := armor.Encode(out, blockType, nil)
	if err != nil {
		return "", err
	}
	_, err = arm.Write(buf.Bytes())
	if err != nil {
		return "", err
	}
	arm.Close()

	return out.String(), nil
}

func PgpEncryptMessageFor(message string, from, to *openpgp.Entity) (string, error) {
	buf := bytes.NewBuffer(nil)
	{
		writer, err := openpgp.Encrypt(buf, []*openpgp.Entity{to}, from, nil, nil)
		if err != nil {
			return "", err
		}
		_, err = writer.Write([]byte(message))
		if err != nil {
			return "", err
		}
		writer.Close()
	}

	out := bytes.NewBuffer(nil)
	{
		arm, err := armor.Encode(out, "PGP MESSAGE", nil)
		if err != nil {
			return "", err
		}
		_, err = arm.Write(buf.Bytes())
		if err != nil {
			return "", err
		}
		arm.Close()
	}
	return out.String(), nil
}
