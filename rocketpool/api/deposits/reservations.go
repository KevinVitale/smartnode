package deposits

import (
    "bytes"
    "encoding/hex"
    "errors"
    "fmt"

    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"
    //"github.com/prysmaticlabs/prysm/shared/ssz"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode-cli/rocketpool/services/accounts"
    "github.com/rocket-pool/smartnode-cli/rocketpool/services/rocketpool"
    "github.com/rocket-pool/smartnode-cli/rocketpool/utils/eth"
)


// DepositInput data
type DepositInput struct {
    pubkey [48]byte
    withdrawalCredentials [32]byte
    proofOfPossession [96]byte
}


// Reserve a node deposit
func reserveDeposit(c *cli.Context, durationId string) error {

    // Initialise account manager
    am := accounts.NewAccountManager(c.GlobalString("keychain"))

    // Get node account
    if !am.NodeAccountExists() {
        fmt.Println("Node account does not exist, please initialize with `rocketpool node init`")
        return nil
    }
    nodeAccount := am.GetNodeAccount()

    // Connect to ethereum node
    client, err := ethclient.Dial(c.GlobalString("provider"))
    if err != nil {
        return errors.New("Error connecting to ethereum node: " + err.Error())
    }

    // Initialise Rocket Pool contract manager
    rp, err := rocketpool.NewContractManager(client, c.GlobalString("storageAddress"))
    if err != nil {
        return err
    }

    // Load Rocket Pool node contracts
    err = rp.LoadContracts([]string{"rocketNodeAPI", "rocketNodeSettings"})
    if err != nil {
        return err
    }
    err = rp.LoadABIs([]string{"rocketNodeContract"})
    if err != nil {
        return err
    }

    // Check node is registered (contract exists)
    nodeContractAddress := new(common.Address)
    err = rp.Contracts["rocketNodeAPI"].Call(nil, nodeContractAddress, "getContract", nodeAccount.Address)
    if err != nil {
        return errors.New("Error checking node registration: " + err.Error())
    }
    if bytes.Equal(nodeContractAddress.Bytes(), make([]byte, common.AddressLength)) {
        fmt.Println("Node is not registered with Rocket Pool, please register with `rocketpool node register`")
        return nil
    }

    // Check node deposits are enabled
    depositsAllowed := new(bool)
    err = rp.Contracts["rocketNodeSettings"].Call(nil, depositsAllowed, "getDepositAllowed")
    if err != nil {
        return errors.New("Error checking node deposits enabled status: " + err.Error())
    }
    if !*depositsAllowed {
        fmt.Println("Node deposits are currently disabled in Rocket Pool")
        return nil
    }

    // Initialise node contract
    nodeContract, err := rp.NewContract(nodeContractAddress, "rocketNodeContract")
    if err != nil {
        return errors.New("Error initialising node contract: " + err.Error())
    }

    // Check node does not have current deposit reservation
    hasReservation := new(bool)
    err = nodeContract.Call(nil, hasReservation, "getHasDepositReservation")
    if err != nil {
        return errors.New("Error retrieving deposit reservation status: " + err.Error())
    }
    if *hasReservation {
        fmt.Println("Node has an existing deposit reservation, please cancel or finalize it")
        return nil
    }

    // Get node's validator pubkey
    // :TODO: implement once BLS library is available
    pubkeyHex := []byte("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcd01")
    pubkey := make([]byte, hex.DecodedLen(len(pubkeyHex)))
    _,_ = hex.Decode(pubkey, pubkeyHex)

    // Get RP withdrawal pubkey
    // :TODO: replace with correct withdrawal pubkey once available
    withdrawalPubkeyHex := []byte("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
    withdrawalPubkey := make([]byte, hex.DecodedLen(len(withdrawalPubkeyHex)))
    _,_ = hex.Decode(withdrawalPubkey, withdrawalPubkeyHex)

    // Build withdrawal credentials
    withdrawalCredentials := eth.KeccakBytes(withdrawalPubkey) // Withdrawal pubkey hash
    withdrawalCredentials[0] = 0 // Replace first byte with BLS_WITHDRAWAL_PREFIX_BYTE

    // Build proof of possession
    // :TODO: implement once BLS library is available
    proofOfPossessionHex := []byte(
        "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef" +
        "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef" +
        "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
    proofOfPossession := make([]byte, hex.DecodedLen(len(proofOfPossessionHex)))
    _,_ = hex.Decode(proofOfPossession, proofOfPossessionHex)

    // Build DepositInput
    // :TODO: implement using SSZ once library is available
    var depositInputLength [4]byte
    depositInputLength[0] = byte(len(pubkey) + len(withdrawalCredentials) + len(proofOfPossession))
    depositInput := bytes.Join([][]byte{depositInputLength[:], pubkey, withdrawalCredentials[:], proofOfPossession}, []byte{})

    /*
    depositInputData := &DepositInput{}
    copy(depositInputData.pubkey[:], pubkey)
    copy(depositInputData.withdrawalCredentials[:], withdrawalCredentials[:])
    copy(depositInputData.proofOfPossession[:], proofOfPossession)
    depositInput := new(bytes.Buffer)
    err = ssz.Encode(depositInput, depositInputData)
    if err != nil {
        return errors.New("Error encoding DepositInput for deposit reservation: " + err.Error())
    }
    */

    // Get node account transactor
    nodeAccountTransactor, err := am.GetNodeAccountTransactor()
    if err != nil {
        return err
    }

    // Create deposit reservation
    _, err = nodeContract.Transact(nodeAccountTransactor, "depositReserve", durationId, depositInput)
    if err != nil {
        return errors.New("Error making deposit reservation: " + err.Error())
    }

    // Log & return
    fmt.Println("Deposit reservation made successfully")
    return nil

}
