# Passw0rd SDK Go

[![Build Status](https://travis-ci.com/VirgilSecurity/virgil-passw0rd-go.png?branch=master)](https://travis-ci.com/VirgilSecurity/virgil-passw0rd-go)
[![GitHub license](https://img.shields.io/badge/license-BSD%203--Clause-blue.svg)](https://github.com/VirgilSecurity/virgil/blob/master/LICENSE)


[Introduction](#introduction) | [Features](#features) | [Register Your Account](#register-your-account) | [Install and configure SDK](#install-and-configure-sdk) | [Prepare Your Database](#prepare-your-database) | [Usage Examples](#usage-examples) | [Docs](#docs) | [Support](#support)

## Introduction
<a href="https://passw0rd.io/"><img width="260px" src="https://cdn.virgilsecurity.com/assets/images/github/logos/passw0rd.png" align="left" hspace="0" vspace="0"></a>[Virgil Security](https://virgilsecurity.com) introduces an implementation of the [Password-Hardened Encryption (PHE) protocol](https://virgilsecurity.com/wp-content/uploads/2018/11/PHE-Whitepaper-2018.pdf) that provides developers with a technology to protect users passwords from offline/online attacks and make stolen passwords useless even if your database has been compromised.

PHE is a new, more secure mechanism that protects user passwords and lessens the security risks associated with weak passwords. Neither Virgil nor attackers know anything about user's password.

**Authors of the PHE protocol**: Russell W. F. Lai, Christoph Egger, Manuel Reinert, Sherman S. M. Chow, Matteo Maffei and Dominique Schroder.

## Features
- Zero knowledge of user password
- Protection from online attacks
- Protection from offline attacks
- Instant invalidation of stolen database
- User data encryption with a personal key


## Register Your Account
Before starting practicing with the SDK and usage examples be sure that:
- you have a registered passw0rd Account
- you created passw0rd Application
- and you got your passw0rd application's credentials, such as: Application Access Token, Service Public Key, Client Secret Key

If you don't have an account or a passw0rd project with its credentials, please use the [passw0rd CLI](https://github.com/passw0rd/cli) to get it.


## Install and Configure SDK
The passw0rd Go SDK is provided as a package named `passw0rd`. The package is distributed via GitHub. The package is available for Go 1.10 or newer.


### Install SDK Package
Install passw0rd SDK library with the following code:
```bash
go get -u github.com/VirgilSecurity/virgil-passw0rd-go
```
Passw0rd Go SDK uses Dep to do manage its dependencies:
Please install [dep](https://golang.github.io/dep/docs/installation.html) and run the following commands:
```bash
cd $(go env GOPATH)/src/github.com/VirgilSecurity/virgil-passw0rd-go
dep ensure
```


### Configure SDK
Here is an example of how to specify your credentials SDK class instance:
```go
// here set your passw0rd credentials
import (
    "github.com/VirgilSecurity/virgil-passw0rd-go"
)

func InitPassw0rd() (*passw0rd.Protocol, error){
    appToken := "PT.OSoPhirdopvijQlFPKdlSydN9BUrn5oEuDwf3Hqps"
    appSecretKey := "SK.1.xacDjofLr2JOu2Vf1+MbEzpdtEP1kUefA0PUJw2UyI0="
    servicePublicKey := "PK.1.BEn/hnuyKV0inZL+kaRUZNvwQ/jkhDQdALrw6VdfvhZhPQQHWyYO+fRlJYZweUz1FGH3WxcZBjA0tL4wn7kE0ls="

    context, err := passw0rd.CreateContext(appToken, servicePublicKey, appSecretKey, "")
    if err != nil{
        return nil, err
    }

    return passw0rd.NewProtocol(context)
}
```



## Prepare Your Database
Passw0rd SDK allows you to easily perform all the necessary operations to create, verify and rotate user's `record`.

**Passw0rd record** - a user's password that is protected with our Passw0rd technology. Passw0rd `record` contains a version, client & server random salts and two values obtained during execution of the PHE protocol.

In order to create and work with user's `record` you have to set up your database with an additional column.

The column must have the following parameters:
<table class="params">
<thead>
		<tr>
			<th>Parameters</th>
			<th>Type</th>
			<th>Size (bytes)</th>
			<th>Description</th>
		</tr>
</thead>

<tbody>
<tr>
	<td>passw0rd_record</td>
	<td>bytearray</td>
	<td>210</td>
	<td> A unique record, namely a user's protected passw0rd.</td>
</tr>

</tbody>
</table>


## Usage Examples

### Enroll User Record

Use this flow to create a new passw0rd's `record` in your DB for a user.

> Remember, if you already have a database with user passwords, you don't have to wait until a user logs in into your system to implement Passw0rd technology. You can go through your database and enroll (create) a user's `record` at any time.

So, in order to create a `record` for a new database or available one, go through the following operations:
- Take a user's **password** (or its hash or whatever you use) and pass it into the `EnrollAccount` function in a SDK on your Server side.
- Passw0rd SDK will send a request to Passw0rd Service to get enrollment.
- Then, Passw0rd SDK will create a user's `record`. You need to store this unique user's `record` in your database in associated column.

```go
package main

import (
    "encoding/base64"
    "fmt"
    "github.com/VirgilSecurity/virgil-passw0rd-go"
    "github.com/VirgilSecurity/virgil-phe-go"
)

// create a new encrypted password record using user password or its hash
func EnrollAccount(password string, prot *passw0rd.Protocol) error{
    
    record, key, err := prot.EnrollAccount(password)
    if err != nil {
        return err
    }

    //save record to database
    fmt.Printf("Database record:\n%s\n", base64.StdEncoding.EncodeToString(record))
    //use encryptionKey for protecting user data
    encrypted, err := phe.Encrypt(data, key)
    ...

}
```

When you've created a passw0rd's `record` for all users in your DB, you can delete the unnecessary column where user passwords were previously stored.


### Verify User Record

Use this flow when a user already has his or her own passw0rd's `record` in your database. This function allows you to verify user's password with the `record` from your DB every time when the user signs in. You have to pass his or her `record` from your DB into the `VerifyPassword` function:

```go
package main

import (
    "fmt"
    "github.com/VirgilSecurity/virgil-passw0rd-go"
    "github.com/VirgilSecurity/virgil-phe-go"
)


func VerifyPassword(password string, record []byte, prot *passw0rd.Protocol) error{
    key, err := prot.VerifyPassword(password, record)
    if err != nil {

        if err == passw0rd.ErrInvalidPassword{
            //invalid password
        }
        return err //some other error
    }

    //use encryptionKey for decrypting user data
    decrypted, err := phe.Decrypt(encrypted, key)
    ...

}
```

## Encrypt user data in your database

Not only user's password is a sensitive data. In this flow we will help you to protect any Personally identifiable information (PII) in your database.

PII is a data that could potentially identify a specific individual, and PII can be sensitive.
Sensitive PII is information which, when disclosed, could result in harm to the individual whose privacy has been breached. Sensitive PII should therefore be encrypted in transit and when data is at rest. Such information includes biometric information, medical information, personally identifiable financial information (PIFI) and unique identifiers such as passport or Social Security numbers.

Passw0rd service allows you to protect user's PII (personal data) with a user's `encryptionKey` that is obtained from `EnrollAccount` or `VerifyPassword` functions. The `encryptionKey` will be the same for both functions.

In addition, this key is unique to a particular user and won't be changed even after rotating (updating) the user's `record`. The `encryptionKey` will be updated after user changes own password.

Here is an example of data encryption/decryption with an `encryptionKey`:

```go
package main

import (
    "fmt"
    "github.com/VirgilSecurity/virgil-phe-go"
)

func main() {

    //key is obtained from proto.EnrollAccount() or proto.VerifyPassword() calls

    data := []byte("Personal data")

    ciphertext, err := phe.Encrypt(data, encryptionKey)
    if err != nil {
        panic(err)
    }
    decrypted, err := phe.Decrypt(ciphertext, encryptionKey)
    if err != nil {
        panic(err)
    }

    //use decrypted data
}
```
Encryption is performed using AES256-GCM with key & nonce derived from the user's encryptionKey using HKDF and random 256-bit salt.

Virgil Security has Zero knowledge about a user's `encryptionKey`, because the key is calculated every time when you execute `EnrollAccount` or `VerifyPassword` functions at your server side.


## Rotate app keys and user record
There can never be enough security, so you should rotate your sensitive data regularly (about once a week). Use this flow to get an `UPDATE_TOKEN` for updating user's passw0rd `RECORD` in your database and to get a new `APP_SECRET_KEY` and `SERVICE_PUBLIC_KEY` of a specific application.

Also, use this flow in case your database has been COMPROMISED!

> This action doesn't require to create an additional table or to do any modification with available one. When a user needs to change his or her own password, use the EnrollAccount function to replace user's oldPassw0rd record value in your DB with a newRecord.

There is how it works:

**Step 1.** Get your `UPDATE_TOKEN` using [Passw0rd CLI](https://github.com/passw0rd/cli)

- be sure you're logged in your account. To log in the account use the following command (2FA is required):

```bash
// FreeBSD / Linux / Mac OS
./passw0rd login my@email.com

// Windows OS
passw0rd login my@email.com
```

- then, use the `rotate` command and your application token to get an `UPDATE_TOKEN`:

```bash
// FreeBSD / Linux / Mac OS
./passw0rd application rotate <app_token>

// Windows OS
passw0rd application rotate <app_token>
```
as a result, you get your `UPDATE_TOKEN`.

**Step 2.** Initialize passw0rd SDK with the `UPDATE_TOKEN`.
Move to passw0rd SDK configuration file and specify your `UPDATE_TOKEN`:

```go
// here set your passw0rd credentials
import (
    "github.com/VirgilSecurity/virgil-passw0rd-go"
)

func InitPassw0rd() (*passw0rd.Protocol, error){
    appToken := "PT.0000000irdopvijQlFPKdlSydN9BUrn5oEuDwf3Hqps"
    appSecretKey := "SK.1.000jofLr2JOu2Vf1+MbEzpdtEP1kUefA0PUJw2UyI0="
    servicePublicKey := "PK.1.BEn/hnuyKV0inZL+kaRUZNvwQ/jkhDQdALrw6Vdf00000QQHWyYO+fRlJYZweUz1FGH3WxcZBjA0tL4wn7kE0ls="
    updateToken := "UT.2.00000000+0000000000000000000008UfxXDUU2FGkMvKhIgqjxA+hsAtf17K5j11Cnf07jB6uVEvxMJT0lMGv00000="

    context, err := passw0rd.CreateContext(appToken, servicePublicKey, appSecretKey, updateToken)
    if err != nil{
        return nil, err
    }

    return passw0rd.NewProtocol(context)
}
```

**Step 3.** Start migration. Use the `NewRecordUpdater("UPDATE_TOKEN")` SDK function to create an instance of class that will update your old records to new ones (you don't need to ask your users to create a new password). The `UpdateRecord()` function requires user's `oldRecord` from your DB:

```go
package main

import (
    "crypto/subtle"
    "github.com/VirgilSecurity/virgil-passw0rd-go"
)

func main(){
	
	updater, err := passw0rd.NewRecordUpdater("UPDATE_TOKEN")
	if err != nil{
            //something went wrong
    }
	
    //for each record
    //get old record from the database
    oldRecord := ...

    //update old record
    newRecord, err := updater.UpdateRecord(oldRecord)
    if err != nil{
        //something went wrong
    }

    // newRecord is nil ONLY if oldRecord is already the latest version
    if newRecord != nil{
        //save new record to the database
        saveNewRecord(newRecord)
    }

}
```

So, run the `UpdateRecord()` function and save user's `newRecord` into your database.

Since the SDK is able to work simultaneously with two versions of user's records (`newRecord` and `oldRecord`), this will not affect the backend or users. This means, if a user logs into your system when you do the migration, the passw0rd SDK will verify his password without any problems because Passw0rd Service can work with both user's records (`newRecord` and `oldRecord`).

**Step 4.** Get a new `APP_SECRET_KEY` and `SERVICE_PUBLIC_KEY` of a specific application

Use passw0rd CLI `update-keys` command and your `UPDATE_TOKEN` to update the `APP_SECRET_KEY` and `SERVICE_PUBLIC_KEY`:

```bash
// FreeBSD / Linux / Mac OS
./passw0rd application update-keys <service_public_key> <app_secret_key> <update_token>

// Windows OS
passw0rd application update-keys <service_public_key> <app_secret_key> <update_token>
```

**Step 5.** Move to passw0rd SDK configuration and replace your previous `APP_SECRET_KEY`,  `SERVICE_PUBLIC_KEY` with a new one (`APP_TOKEN` will be the same). Delete previous `APP_SECRET_KEY`, `SERVICE_PUBLIC_KEY` and `UPDATE_TOKEN`.

```go
// here set your passw0rd credentials
import (
    "github.com/VirgilSecurity/virgil-passw0rd-go"
)

func InitPassw0rd() (*passw0rd.Protocol, error){
    appToken := "APP_TOKEN_HERE"
    appSecretKey := "NEW_APP_SECRET_KEY_HERE"
    servicePublicKey := "NEW_SERVICE_PUBLIC_KEY_HERE"


    context, err := passw0rd.CreateContext(appToken, servicePublicKey, appSecretKey, "")
    if err != nil{
        return nil, err
    }

    return passw0rd.NewProtocol(context)
}
```



## Docs
* [Passw0rd][_passw0rd] home page
* [The PHE WhitePaper](https://virgilsecurity.com/wp-content/uploads/2018/11/PHE-Whitepaper-2018.pdf) - foundation principles of the protocol

## License

This library is released under the [3-clause BSD License](LICENSE.md).

## Support
Our developer support team is here to help you. Find out more information on our [Help Center](https://help.virgilsecurity.com/).

Also, get extra help from our support team: support@VirgilSecurity.com.

[_passw0rd]: https://passw0rd.io/
