package main

import (
	"bytes"
	"os"

	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/armor"
	"golang.org/x/crypto/openpgp/packet"
)

func readReadKey(fileName string) *openpgp.Entity {
	reader, err := os.Open(fileName)
	if err != nil {
		return nil
	}
	defer reader.Close()

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

func writeEntity(entity *openpgp.Entity, private bool, fileName string) error {
	blockType := "PGP PUBLIC KEY BLOCK"
	if private {
		blockType = "PGP PRIVATE KEY BLOCK"
	}

	writer, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer writer.Close()

	buf := bytes.NewBuffer(nil)
	if private {
		err = entity.SerializePrivate(buf, nil)
	} else {
		err = entity.Serialize(buf)
	}
	if err != nil {
		return err
	}

	out := bytes.NewBuffer(nil)
	arm, err := armor.Encode(out, blockType, nil)
	if err != nil {
		return err
	}
	_, err = arm.Write(buf.Bytes())
	if err != nil {
		return err
	}
	arm.Close()

	_, err = writer.Write(out.Bytes())
	if err != nil {
		return err
	}
	_, err = writer.WriteString("\n")
	if err != nil {
		return err
	}
	return nil
}

func writePublicKey(entity *openpgp.Entity, fileName string) error {
	return writeEntity(entity, false, fileName)
}

func writePrivateKey(entity *openpgp.Entity, fileName string) error {
	return writeEntity(entity, true, fileName)
}

func loadCondensatKey() *openpgp.Entity {
	pgpFile := "condensat.asc"
	pgpPubFile := "condensat_pub.asc"
	condensat := readReadKey(pgpFile)
	if condensat == nil {
		var err error
		condensat, err = openpgp.NewEntity(
			"CondensatBank",
			"Condensat PGP identity",
			"bank@condensat.tech",
			nil)
		if err != nil {
			panic(err)
		}
		err = writePrivateKey(condensat, pgpFile)
		if err != nil {
			panic(err)
		}
		err = writePublicKey(condensat, pgpPubFile)
		if err != nil {
			panic(err)
		}
	}

	return condensat
}

func test_pgp() {
	condensat := loadCondensatKey()
	if condensat == nil {
		panic("No Keys")
	}
	pubKeyFile := "marsu.asc"

	client := readReadKey(pubKeyFile)

	buf := bytes.NewBuffer(nil)
	writer, err := openpgp.Encrypt(buf, []*openpgp.Entity{client}, condensat, nil, nil)
	if err != nil {
		panic(err)
	}
	defer writer.Close()
	_, err = writer.Write([]byte("Hello, New Customer!\n"))
	if err != nil {
		panic(err)
	}
	writer.Close()

	out := bytes.NewBuffer(nil)
	arm, err := armor.Encode(out, "PGP MESSAGE", nil)
	if err != nil {
		panic(err)
	}
	_, err = arm.Write(buf.Bytes())
	if err != nil {
		panic(err)
	}
	arm.Close()

	os.Stdout.Write(out.Bytes())
	os.Stdout.Write([]byte("\n"))
}
