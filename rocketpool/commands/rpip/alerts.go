package rpip

import (
    "github.com/rocket-pool/smartnode-cli/rocketpool/utils/storage"
)


// Subscribe to alerts
func AlertsSubscribe(email string) error {

    // Open storage
    store, err := storage.Open();
    if err != nil {
        return err
    }
    defer store.Close();

    // Store email address
    return store.Put("alert.subscription.address", email)

}


// Get subscribed address
func AlertsGetSubscribed() (string, error) {

    // Open storage
    store, err := storage.Open();
    if err != nil {
        return "", err
    }
    defer store.Close();

    // Get email address
    var email string = ""
    store.Get("alert.subscription.address", &email)
    return email, nil

}


// Unsubscribe from alerts
func AlertsUnsubscribe() error {

    // Open storage
    store, err := storage.Open();
    if err != nil {
        return err
    }
    defer store.Close();

    // Delete email address
    store.Delete("alert.subscription.address")
    return nil

}
