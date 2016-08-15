/*
Copyright IBM Corp 2016 All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
ou may not use this file except in compliance with the License.
You may obtain a copy of the License at

		 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	//"strings"
	"time"
	//"reflect"
	"math/rand"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

//Account account of user who can vote
type Account struct {
	ID         string   `json:"account_id"`
	Name       string   `json:"name"`
	VoteCount  uint64   `json:"vote_count"`
	Email      string   `json:"email"`
	Org        string   `json:"org"`
	ReqTime    string   `json:"req_time"`
	Privileges []string `json:"privileges"`
}

var accountHeader = "account::"

//Topic voting topic and choices
type Topic struct {
	ID         string   `json:"topic_id"`
	TopicStr   string   `json:"topic"`
	Issuer     string   `json:"issuer"`
	Choices    []string `json:"choices[]"`
	Votes      []string `json:"votes[]"` //ints in string form
	IssueDate  string   `json:"issue_date"`
	ExpireDate string   `json:"expire_date"`
}

var topicHeader = "topic::"

//Vote vote cast for a given topic
type Vote struct {
	Topic    string   `json:"topic"` //topic being voted upon
	Voter    string   `json:"voter"`
	CastDate string   `json:"castDate"` //current time as a string
	Choices  []string `json:"choices[]"`
	Votes    []string `json:"votes[]"`
}

var voteHeader = "vote::"

// ============================================================================================================================
// Main
// ============================================================================================================================
func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

func (t *SimpleChaincode) readStringSafe(col *shim.Column) string {
	if col == nil {
		return ""
	}

	return col.GetString_()
}

func (t *SimpleChaincode) readInt64Safe(col *shim.Column) int64 {
	if col == nil {
		return 0
	}

	return col.GetInt64()
}

func (t *SimpleChaincode) readUint64Safe(col *shim.Column) uint64 {
	if col == nil {
		return 0
	}

	return col.GetUint64()
}

func (t *SimpleChaincode) readBoolSafe(col *shim.Column) bool {
	if col == nil {
		return false
	}

	return col.GetBool()
}

func (t *SimpleChaincode) read(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	var name string
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1: name of the var to query")
	}

	fmt.Println("ARGS", args)

	name = args[0]
	valAsbytes, err := stub.GetState(name)
	if err != nil {
		return nil, errors.New("Error: failed to get state for " + name)
	}

	fmt.Println("BYTES", valAsbytes)

	return valAsbytes, nil
}

func (t *SimpleChaincode) write(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	var name, value string
	var err error
	fmt.Println("running write")

	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2: name of the variable and value to set")
	}

	name = args[0]
	value = args[1]
	err = stub.PutState(name, []byte(value))
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (t *SimpleChaincode) checkAccount(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	fmt.Println("inside check account args")
	fmt.Println(args)
	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2")
	}

	email := args[0]
	userID := args[1]
	if userID == "master-manager" && email == "manager" {
		fmt.Println("Got Manager")
		return nil, nil
	}
	//find if account exist or not
	var column []shim.Column
	column = append(column, shim.Column{Value: &shim.Column_String_{String_: userID}})
	row, errGetRow := stub.GetRow("ApprovedAccounts", column)
	if len(row.Columns) == 0 || errGetRow != nil || t.readStringSafe(row.Columns[2]) != email {
		fmt.Println("UserID does not exist. Please click on forgot password to recover account. [Just kidding, you can't]")
		return nil, errors.New("UserID already exist. Please click on forgot password to recover account")
	}
	return nil, nil
}

func (t *SimpleChaincode) requestAccount(stub *shim.ChaincodeStub, args []string) ([]byte, error) {

	if len(args) != 4 {
		fmt.Println("Not enough arguments passed to createAcount")
		return nil, errors.New("Incorrect number of arguments. Expecting 3: username, email, org, privileges")
	}

	fmt.Println("In request Account username= " + args[0] + " email = " + args[1] + " org = " + args[2] + "privileges = " + args[3])
	var account = Account{ID: "", Name: args[0], Email: args[1], VoteCount: 0, Org: args[2], Privileges: strings.Split(args[3], ",")}
	//Check if request already exists
	//user email must be unique in all the accounts

	var column []shim.Column
	column = append(column, shim.Column{Value: &shim.Column_String_{String_: account.Email}})
	row, errGetRow := stub.GetRow("AccountRequests", column)

	if len(row.Columns) != 0 || errGetRow != nil {
		fmt.Println("Email ID [%s] already exist. Please click on forgot password to recover account. ERR: [%s]", account.Email, errGetRow)
		return nil, fmt.Errorf("Email ID [%s] already exist. Please click on forgot password to recover account. ERR: [%s]", account.Email, errGetRow)
	}
	//TODO check if row does not exist then only execute this code
	//right npw fmt.println(row) is giving {[]} but its not giving any error so ask dale.
	/*if !row {
		return nil, fmt.Errorf("Email ID [%s] already exist. Please click on forgot password to recover account. ERR: [%s]", account.Email, errGetRow)
	}
	*/
	//Account does not exists
	requestTime := time.Now().Format("02 Jan 06 15:04 MST")
	fmt.Println(requestTime)
	rowAdded, rowErr := stub.InsertRow("AccountRequests", shim.Row{
		Columns: []*shim.Column{
			&shim.Column{Value: &shim.Column_String_{String_: account.Email}},
			&shim.Column{Value: &shim.Column_String_{String_: account.Name}},
			&shim.Column{Value: &shim.Column_String_{String_: "open"}},
			&shim.Column{Value: &shim.Column_String_{String_: account.Org}},
			&shim.Column{Value: &shim.Column_String_{String_: strings.Join(account.Privileges, ",")}},
			&shim.Column{Value: &shim.Column_String_{String_: requestTime}},
		},
	})
	if rowErr != nil || !rowAdded {
		fmt.Println(fmt.Sprintf("[ERROR] Could not insert a message into the ledger mostly because email is already registered: %s", rowErr))
		return nil, nil
	}

	row, errGetRow = stub.GetRow("AccountRequests", column)
	if len(row.Columns) != 0 && errGetRow == nil {
		fmt.Println("Row is added in request account")
		fmt.Println(row)
		return nil, nil
	}

	return nil, errors.New("Can not add row in account request table")
}

func (t *SimpleChaincode) getUserID(stub *shim.ChaincodeStub, args []string) (string, error) {
	email := args[0]
	//fmt.Println("In get User ID args:")
	//fmt.Println(args)
	rowChan, rowErr := stub.GetRows("ApprovedAccounts", []shim.Column{})
	if rowErr != nil {
		fmt.Println(fmt.Sprintf("[ERROR] Could not retrieve the rows: %s", rowErr))
		return "", rowErr
	}
	fmt.Println("in get User ID chanValue:")
	for chanValue := range rowChan {
		//fmt.Println(chanValue);
		if t.readStringSafe(chanValue.Columns[2]) == email {
			return t.readStringSafe(chanValue.Columns[0]), nil
		}
	}
	return "", errors.New("Can not find email. Are you sure you are registred?")
}

// getAccount returns the account matching the given username

func (t *SimpleChaincode) getAccount(stub *shim.ChaincodeStub, args []string) (Account, error) {

	var account Account
	var err error
	//fmt.Println("before get user id args are")
	//fmt.Println(args)
	account.ID, err = t.getUserID(stub, args)

	if err != nil {
		return Account{}, err
	}
	var column []shim.Column
	column = append(column, shim.Column{Value: &shim.Column_String_{String_: account.ID}})
	row, errGetRow := stub.GetRow("ApprovedAccounts", column)
	if len(row.Columns) == 0 || errGetRow != nil {
		fmt.Println("UserID does not exist. Please click on forgot password to recover account. [Just kidding]")
		return Account{}, errors.New("UserID already exist. Please click on forgot password to recover account")
	}

	account.Name = t.readStringSafe(row.Columns[1])
	account.Email = args[0]
	account.Org = t.readStringSafe(row.Columns[3])
	account.Privileges = strings.Split(t.readStringSafe(row.Columns[4]), ",")
	account.VoteCount = t.readUint64Safe(row.Columns[5])
	account.ReqTime = t.readStringSafe(row.Columns[6])

	//account.ID = "******" //blank out account ID so user cannot view it
	fmt.Println("get account return value")
	fmt.Println(account)

	return account, nil
}

func (t *SimpleChaincode) getOpenRequests(stub *shim.ChaincodeStub) ([]Account, error) {

	rowChan, rowErr := stub.GetRows("AccountRequests", []shim.Column{})
	if rowErr != nil {
		fmt.Println(fmt.Sprintf("[ERROR] Could not retrieve the rows: %s", rowErr))
		return nil, rowErr
	}
	var openRequest []Account
	for chanValue := range rowChan {
		if chanValue.Columns[2].GetString_() == "open" {
			openRequest = append(openRequest, Account{
				Email:      chanValue.Columns[0].GetString_(),
				Name:       chanValue.Columns[1].GetString_(),
				VoteCount:  0,
				Org:        chanValue.Columns[3].GetString_(),
				Privileges: strings.Split(chanValue.Columns[4].GetString_(), ","),
				ReqTime:    chanValue.Columns[5].GetString_(),
			})
			//timings = append(timings, chanValue.Columns[4].GetString_())
		}
	}
	return openRequest, nil
}

func (t *SimpleChaincode) replaceRowRequest(stub *shim.ChaincodeStub, args []string) (string, error) {
	status := args[0]
	//votes, _ := strconv.ParseUint(args[2], 10, 64)
	account := Account{Name: args[1], Email: args[2]}

	//getrow to save request time before deleting
	fmt.Println("Account Email inside replece row")
	fmt.Println(account.Email)
	var requestTime string

	var column []shim.Column
	column = append(column, shim.Column{Value: &shim.Column_String_{String_: account.Email}})
	row, errGetRow := stub.GetRow("AccountRequests", column)

	if errGetRow != nil || len(row.Columns) == 0 {
		fmt.Println(fmt.Sprintf("[ERROR] Could not retrieve the rows: %s", errors.New("Failed to find row")))
		return "a", errors.New("Failed to find row")
	}
	fmt.Println("In replace row:")
	fmt.Println(row)
	fmt.Println(t.readStringSafe(row.Columns[5]))
	requestTime = t.readStringSafe(row.Columns[5])
	fmt.Println("request time = " + requestTime)
	//Delete old row
	err := stub.DeleteRow(
		"AccountRequests",
		[]shim.Column{shim.Column{Value: &shim.Column_String_{String_: account.Email}}},
	)
	if err != nil {
		return "a", errors.New("Failed deliting row.")
	}

	//inster new row with new status
	_, err = stub.InsertRow("AccountRequests",
		shim.Row{
			Columns: []*shim.Column{
				&shim.Column{Value: &shim.Column_String_{String_: account.Email}},
				&shim.Column{Value: &shim.Column_String_{String_: account.Name}},
				&shim.Column{Value: &shim.Column_String_{String_: status}},
				&shim.Column{Value: &shim.Column_String_{String_: account.Org}},
				&shim.Column{Value: &shim.Column_String_{String_: strings.Join(account.Privileges, ",")}},
				&shim.Column{Value: &shim.Column_String_{String_: requestTime}},
			},
		})
	if err != nil {
		return "a", errors.New("Failed inserting row.")
	}
	return requestTime, nil
}

func generateUserID() string {
	//random number generator
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	const (
		letterIdxBits = 6                    // 6 bits to represent a letter index
		letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
		letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
	)
	n := 6
	var src = rand.NewSource(time.Now().UnixNano())
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}
	fmt.Println("Randmly generated String:")
	fmt.Println(string(b))
	return string(b)
}

func (t *SimpleChaincode) changeStatus(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	fmt.Println("Inside change status args are: ")
	fmt.Println(args)
	status := args[0]
	account := Account{Name: args[1], Email: args[2], Org: args[3], Privileges: strings.Split(args[4], ",")}
	reqTime, errReplceRow := t.replaceRowRequest(stub, args)
	if errReplceRow != nil {
		return nil, errReplceRow
	}

	var userID string
	if status == "approved" {
		userID = generateUserID()
		fmt.Println("My random ID is:")
		fmt.Println(userID)
		manager := args[5]
		votes, _ := strconv.ParseUint(args[6], 10, 64)
		_, err := stub.InsertRow("ApprovedAccounts",
			shim.Row{
				Columns: []*shim.Column{
					&shim.Column{Value: &shim.Column_String_{String_: userID}},
					&shim.Column{Value: &shim.Column_String_{String_: account.Name}},
					&shim.Column{Value: &shim.Column_String_{String_: account.Email}},
					&shim.Column{Value: &shim.Column_String_{String_: account.Org}},
					&shim.Column{Value: &shim.Column_String_{String_: strings.Join(account.Privileges, ",")}},
					&shim.Column{Value: &shim.Column_Uint64{Uint64: votes}},
					&shim.Column{Value: &shim.Column_String_{String_: reqTime}},
					&shim.Column{Value: &shim.Column_String_{String_: time.Now().String()}},
					&shim.Column{Value: &shim.Column_String_{String_: manager}},
				},
			})
		if err != nil {
			return nil, errors.New("Failed inserting row.")
		}
		rowChan, rowErr := stub.GetRows("ApprovedAccounts", []shim.Column{})
		if rowErr != nil {
			fmt.Println(fmt.Sprintf("[ERROR] Could not retrieve the rows: %s", rowErr))
			return nil, rowErr
		}
		fmt.Println("chanValue:")
		for chanValue := range rowChan {
			fmt.Println(chanValue.Columns[1])
		}
	}
	return nil, nil
}

func (t *SimpleChaincode) getAllRequests(stub *shim.ChaincodeStub, accountID string) (Account, error) {
	var account Account
	accountBytes, err := stub.GetState(accountHeader + accountID)
	if err != nil {
		fmt.Println("Could not find account " + accountID)
		return account, err
	}

	err = json.Unmarshal(accountBytes, &account)
	if err != nil {
		fmt.Println("Error unmarshalling account " + accountID + "\n err: " + err.Error())
		return account, err
	}

	return account, nil
}

func (t *SimpleChaincode) issueTopic(stub *shim.ChaincodeStub, args []string) ([]byte, error) {

	fmt.Println("in issue topic args")
	fmt.Println(args)

	if len(args) != 1 {
		fmt.Println("Incorrect number of arguments. Expecting 1: json object of topic being issued")
		return nil, errors.New("Incorrect number of arguments. Expecting 1: json object of topic being issued")
	}

	var topic Topic
	var err error
	//var account Account

	fmt.Println("Unmarshalling topic")
	err = json.Unmarshal([]byte(args[0]), &topic)
	if err != nil {
		fmt.Println("Invalid topic issued")
		return nil, err
	}

	fmt.Println("Getting state of issuer " + topic.Issuer)
	/*
		accountBytes, err := stub.GetState(accountHeader + topic.Issuer)
		if err != nil {
			fmt.Println("Error getting state of - " + topic.Issuer)
			return nil, err
		}
		err = json.Unmarshal(accountBytes, &account)
		if err != nil {
			fmt.Println("Error unmarshalling accountBytes")
			return nil, err
		}
	*/
	fmt.Println("Getting state on topic " + topic.TopicStr)
	existingTopicBytes, err := stub.GetState(topicHeader + topic.ID)
	if existingTopicBytes == nil {
		fmt.Println("Topic does not exist, creating new topic...")

		//create empty array of votes in topic length of choices
		topic.Votes = make([]string, len(topic.Choices))
		for i := 0; i < len(topic.Votes); i++ {
			topic.Votes[i] = "0"
		}

		//change expire_date to go time format
		dur, err := time.ParseDuration(topic.ExpireDate + "ms")
		if err != nil {
			fmt.Println(err)
			return nil, err
		}

		topic.ExpireDate = (time.Now().Add(dur)).Format(time.RFC3339)
		fmt.Println(topic.ExpireDate)

		//set issue_date to current time
		issueDateTime := time.Now()
		topic.IssueDate = issueDateTime.Format(time.RFC3339)

		topicBytes, err := json.Marshal(&topic)
		if err != nil {
			fmt.Println("Error marshalling topic")
			return nil, err
		}

		err = stub.PutState(topicHeader+topic.ID, topicBytes)
		if err != nil {
			fmt.Println("Error issuing topic " + topic.TopicStr)
			return nil, err
		}

		fmt.Println("Getting Vote Topics")
		voteTopicsBytes, err := stub.GetState("VoteTopics")
		if err != nil {
			fmt.Println("Error retrieving Vote Topics")
			return nil, err
		}
		var voteTopics []string
		err = json.Unmarshal(voteTopicsBytes, &voteTopics)
		if err != nil {
			fmt.Println("Error unmarshalling Vote Topics")
			return nil, err
		}

		fmt.Println("Appending the new topic to Vote Topics")
		foundTopic := false
		for _, tmp := range voteTopics {
			if tmp == topic.ID {
				foundTopic = true
			}
		}
		if foundTopic == false {
			voteTopics = append(voteTopics, topic.ID)
			voteTopicBytesToWrite, err := json.Marshal(&voteTopics)
			if err != nil {
				fmt.Println("Error marshalling vote topics")
				return nil, err
			}
			fmt.Println("Put state on Vote Topics")
			err = stub.PutState("VoteTopics", voteTopicBytesToWrite)
			if err != nil {
				fmt.Println("Error writting vote topics back")
				return nil, err
			}
		}

		//getting here means success so far
		//create table associated with topic
		fmt.Println("CREATING TABLE", topicHeader+topic.ID)
		errCreateTable := stub.CreateTable(topicHeader+topic.ID, []*shim.ColumnDefinition{
			&shim.ColumnDefinition{Name: "TransactionID", Type: shim.ColumnDefinition_UINT64, Key: true},
			&shim.ColumnDefinition{Name: "Voter", Type: shim.ColumnDefinition_STRING, Key: false},
			&shim.ColumnDefinition{Name: "Choice", Type: shim.ColumnDefinition_STRING, Key: false},
			&shim.ColumnDefinition{Name: "Votes", Type: shim.ColumnDefinition_UINT64, Key: false},
			&shim.ColumnDefinition{Name: "Time", Type: shim.ColumnDefinition_STRING, Key: false},
		})

		if errCreateTable != nil {
			fmt.Println("Error creating topic "+topic.TopicStr+" table: ", errCreateTable)
			return nil, errCreateTable
		}

		//all success
		fmt.Println("Issued topic " + topic.TopicStr)
		return nil, nil
	}

	fmt.Println("Topic already exists")
	return nil, nil
}

//ClearTopics is for debugging to clear all topics on ledger
func (t *SimpleChaincode) clearTopics(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	fmt.Println("Clearing all topics...")

	topics, err := t.getAllTopics(stub)
	if err != nil {
		fmt.Println("Error: Could not retrieve voting topics: ", err)
		return nil, err
	}

	for _, topic := range topics {
		fmt.Println("Clearing topic ID \"" + topic.TopicStr + "\"...")

		err2 := stub.DelState(topicHeader + topic.ID)
		if err2 != nil {
			fmt.Println("Error: Failed to clear vote topic \""+topic.TopicStr+"\": ", err2)
			return nil, err2
		}
		fmt.Println("Successfully cleared vote topic ID " + topic.TopicStr)

		err2 = stub.DeleteTable(topicHeader + topic.ID)
		if err2 != nil {
			fmt.Println("Error: Failed to delete table for vote topic \""+topic.TopicStr+"\": ", err2)
			return nil, err2
		}
	}

	var blank []string
	blankBytes, _ := json.Marshal(&blank)
	err2 := stub.PutState("VoteTopics", blankBytes)
	if err2 != nil {
		fmt.Println("Error: Failed to clear vote topics: ", err2)
		return nil, err2
	}
	fmt.Println("Successfully cleared vote topics")
	return nil, nil
}

//getAllTopics returns an array of all topicIDs
func (t *SimpleChaincode) getAllTopics(stub *shim.ChaincodeStub) ([]Topic, error) {
	fmt.Println("Retrieving all topics...")

	var allTopics []Topic

	topicsBytes, err := stub.GetState("VoteTopics")
	if err != nil {
		fmt.Println("Error retrieving vote topics")
		return nil, err
	}

	var topics []string
	err = json.Unmarshal(topicsBytes, &topics)
	if err != nil {
		fmt.Println("Error unmarshalling vote topics: ", err)
		return nil, err
	}

	for _, value := range topics {
		topicBytes, err := stub.GetState(topicHeader + value)
		if err != nil {
			fmt.Println("Error retrieving topic "+value+": ", err)
			return nil, err
		}

		var topic Topic
		err = json.Unmarshal(topicBytes, &topic)
		if err != nil {
			fmt.Println("Error unmarshalling topic "+value+": ", err)
			return nil, err
		}

		fmt.Println("Appending topic " + topic.TopicStr)
		allTopics = append(allTopics, topic)
	}

	return allTopics, nil
}

func (t *SimpleChaincode) getTopic(stub *shim.ChaincodeStub, topicID string) (Topic, error) {
	var emptyTopic Topic

	fmt.Println("Retrieving topic " + topicID + "...")

	topicBytes, err := stub.GetState(topicHeader + topicID)
	if err != nil {
		fmt.Println("Error retrieving vote topic")
		return emptyTopic, err
	}

	//fmt.Println(topicBytes)

	var topic Topic
	err = json.Unmarshal(topicBytes, &topic)
	if err != nil {
		fmt.Println("Error unmarshalling vote topic "+topicID+": ", err)
		return emptyTopic, err
	}

	return topic, nil
}

var transactionID uint64

func (t *SimpleChaincode) hasUserVoted(stub *shim.ChaincodeStub, args []string) (bool, error) {

	if len(args) != 2 {
		fmt.Println("Incorrect number of arguments. Expecting 2: topic ID and account ID")
		return true, errors.New("Incorrect number of arguments. Expecting 2: topic ID and account ID")
	}

	fmt.Println("Checking whether user " + args[1] + " has voted on topic " + args[0] + "...")

	topicID := args[0]
	email := args[1]

	account, errGetAccount := t.getAccount(stub, []string{email})
	if errGetAccount != nil {
		fmt.Println("Error retrieving account: ", errGetAccount)
		return true, errGetAccount
	}

	topicBytes, errTopic := stub.GetState(topicHeader + topicID)
	if errTopic != nil {
		fmt.Println("[ERROR] Error retrieving topic "+topicID+": ", errTopic)
		return true, errTopic
	}

	var topic Topic
	errJSON := json.Unmarshal(topicBytes, &topic)
	if errJSON != nil {
		fmt.Println("[ERROR] Error unmarshalling topic "+topicID+": ", errJSON)
		return true, errJSON
	}

	table, err := stub.GetTable(topicHeader + topic.ID)
	fmt.Println(table)
	fmt.Println(err)

	rowChan, rowErr := stub.GetRows(topicHeader+topic.ID, []shim.Column{})
	if rowErr != nil {
		fmt.Println(fmt.Sprintf("[ERROR] Could not retrieve the rows: %s", rowErr))
		return true, rowErr
	}

	for chanValue := range rowChan {
		fmt.Println(chanValue)
		if t.readStringSafe(chanValue.Columns[1]) == account.Email {
			return true, nil //user has voted!
		}
	}

	return false, nil
}

func (t *SimpleChaincode) castVote(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	/*		0
			json
			{
				"topic_id": "string",
				"voter": "username",
				"votes": [option1, option2, ...]
			}
	*/

	if len(args) != 1 {
		fmt.Println("Incorrect number of arguments. Expecting 1: json object of vote being cast")
		return nil, errors.New("Incorrect number of arguments. Expecting 1: json object of vote being cast")
	}

	var vote Vote

	fmt.Println("Unmarshalling vote")
	err := json.Unmarshal([]byte(args[0]), &vote)
	if err != nil {
		fmt.Println("Invalid vote cast: ", err)
		return nil, err
	}

	fmt.Println("Vote: ", vote)

	account, errGetAccount := t.getAccount(stub, []string{vote.Voter})

	if errGetAccount != nil {
		fmt.Println("Error retrieving account: ", errGetAccount)
		return nil, errGetAccount
	}

	topicBytes, errTopic := stub.GetState(topicHeader + vote.Topic)
	if errTopic != nil {
		fmt.Println("Error retrieving topic "+vote.Topic+": ", errTopic)
		return nil, errTopic
	}

	var topic Topic
	errJSON := json.Unmarshal(topicBytes, &topic)
	if errJSON != nil {
		fmt.Println("Error unmarshalling topic "+vote.Topic+": ", errJSON)
		return nil, errJSON
	}

	//check votes are valid

	//make sure topic has not expired
	expireTime, errTimeParse := time.Parse(time.RFC3339, topic.ExpireDate)
	if errTimeParse != nil {
		fmt.Println(errTimeParse)
		return nil, errTimeParse
	}
	if !(time.Now().Before(expireTime)) {
		fmt.Println("[ERROR] Attempted to cast vote on expired topic")
		return nil, errors.New("Attempted to cast vote on expired topic")
	}

	//make sure all votes are >=0
	var count uint64
	for _, quantityStr := range vote.Votes {
		quantity, err := strconv.Atoi(quantityStr)
		if err != nil {
			fmt.Println("Error converting vote from string to int: ", err)
			return nil, err
		}
		if quantity < 0 {
			fmt.Println("Error: attempted to cast vote of negative value")
			return nil, errors.New("Attempted to cast vote of negative value")
		}
		count += uint64(quantity)
	}

	//make sure voter has not cast more votes than allowed
	if count > account.VoteCount {
		fmt.Println("Error: attempted to cast more votes than voter has")
		return nil, errors.New("Attempted to cast more votes than voter has")
	}

	//make sure voter has cast correct number of votes
	if len(vote.Votes) != len(topic.Choices) {
		fmt.Println("Error: number of vote quantities (" + strconv.Itoa(len(vote.Votes)) + ") does not match choices count (" + strconv.Itoa(len(topic.Choices)) + ")")
		return nil, errors.New("Number of vote quantities (" + strconv.Itoa(len(vote.Votes)) + ") does not match choices count (" + strconv.Itoa(len(topic.Choices)) + ")")
	}

	fmt.Println("Casting votes for topic " + topic.TopicStr + "...")

	for i := 0; i < len(topic.Choices); i++ {
		fmt.Println("Casting vote for choice " + topic.Choices[i])
		voteQty, err := strconv.Atoi(vote.Votes[i])
		fmt.Println("Vote quantity: ", voteQty)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		if voteQty > 0 {
			//add to array in Topic
			topicVoteTally, err := strconv.Atoi(topic.Votes[i])
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			topic.Votes[i] = strconv.Itoa(topicVoteTally + voteQty) //convery to int, add vote, then convert back to string

			//add to table

			fmt.Println("ADDING ROW", topicHeader+vote.Topic)
			addedRow, errRow := stub.InsertRow(topicHeader+vote.Topic, shim.Row{
				Columns: []*shim.Column{
					{&shim.Column_Uint64{Uint64: transactionID}},
					{&shim.Column_String_{String_: vote.Voter}},
					{&shim.Column_String_{String_: topic.Choices[i]}},
					{&shim.Column_Uint64{Uint64: uint64(voteQty)}},
					{&shim.Column_String_{String_: vote.CastDate}},
				},
			})

			fmt.Println("[ADDED ROW]", addedRow)

			rowChan, rowErr := stub.GetRows(topicHeader+vote.Topic, []shim.Column{})
			if rowErr != nil {
				fmt.Println(fmt.Sprintf("[ERROR] Could not retrieve the rows: %s", rowErr))
				return nil, rowErr
			}
			for chanValue := range rowChan {
				fmt.Println("CHAN VAL", chanValue)
			}

			if errRow != nil || !addedRow {
				fmt.Println("Error creating row in table "+vote.Topic+": ", errRow)
				return nil, errRow
			}

			transactionID++
		}
	}

	//rewrite topic
	topicBytes, err2 := json.Marshal(&topic)
	if err2 != nil {
		fmt.Println(err2)
		return nil, err2
	}
	err2 = stub.PutState(topicHeader+topic.ID, topicBytes)
	if err2 != nil {
		fmt.Println(err2)
		return nil, err2
	}

	fmt.Println("Vote successfully cast!")

	return nil, nil
}

func (t *SimpleChaincode) tallyVotes(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	if len(args) != 1 {
		fmt.Println("Incorrect number of arguments. Expecting 1: string of topic ID to be queried")
		return nil, errors.New("Incorrect number of arguments. Expecting 1: string of topic ID to be queried")
	}
	/*
		_, errGetAccount := getTopic(stub, args[0])
		if errGetAccount != nil {
			fmt.Println("Could not retrieve vote topic to be tallied")
			return nil, errGetAccount
		}

				for _, col := range row.GetColumns() {
					fmt.Println("[INFO] Column: ", col)
				}
				fmt.Println(fmt.Sprintf("[INFO] Row: %v", row))
			}
		}
	*/
	return nil, nil
}

// Invoke is our entry point to invoke a chaincode function
func (t *SimpleChaincode) Invoke(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	fmt.Println("invoke is running " + function)
	fmt.Println(args)
	// Handle different functions
	switch function {
	case "init": //initialize the chaincode state, used as reset
		return t.Init(stub, "init", args)
	case "write":
		return t.write(stub, args)
	case "issue_topic":
		return t.issueTopic(stub, args)
	case "clear_all_topics":
		return t.clearTopics(stub, args)
	case "request_account":
		return t.requestAccount(stub, args)
	case "change_status":
		return t.changeStatus(stub, args)
	case "cast_vote":
		return t.castVote(stub, args)
	}

	fmt.Println("invoke did not find func: " + function) //error

	return nil, errors.New("Received unknown function invocation")
}

// Query is our entry point for queries
func (t *SimpleChaincode) Query(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	fmt.Println("query is running " + function)
	//fmt.Println(args)
	// Handle different functions
	switch function {

	case "check_account":
		a, err := t.checkAccount(stub, args)
		return a, err

	case "get_UserID":
		userID, err := t.getUserID(stub, args)

		if err != nil {
			return nil, err
		}

		//change these names
		//they are no good
		type JSONcapsule struct {
			AllAccReq string
		}
		AccReqJSON := JSONcapsule{
			AllAccReq: userID,
		}
		fmt.Println("AccReqJSON")
		fmt.Println(AccReqJSON)
		userIDJSON, err1 := json.Marshal(&AccReqJSON)
		if err1 != nil {
			return nil, err1
		}
		return userIDJSON, nil

	case "read": //read a variable
		return t.read(stub, args)

		// type Capsule struct {
		// 	Data []byte `json:"data"`
		// }

		// var capsule Capsule
		// data, err := t.read(stub, args)

		// if err != nil {
		// 	fmt.Println(err)
		// 	return nil, err
		// }

		// capsule.Data = data
		// dataBytes, err1 := json.Marshal(&capsule)
		// if err1 != nil {
		// 	fmt.Println("Error marshalling read data")
		// 	return nil, err1
		// }
		// return dataBytes, nil

	case "get_all_topics":
		if len(args) != 1 {
			fmt.Println("Incorrect number of arguments. Expecting 1: user ID")
			return nil, errors.New("Incorrect number of arguments. Expecting 1: user ID")
		}

		account, errAccount := t.getAccount(stub, []string{args[0]})
		fmt.Println("in get all topics after get account account variable is:")
		fmt.Println(account)

		if errAccount != nil {
			fmt.Println("Error getting account:", errAccount)
			return nil, errAccount
		}

		allTopics, err := t.getAllTopics(stub)
		fmt.Println("in get all topics after get all topics topics are")
		fmt.Println(allTopics)

		if err != nil {
			fmt.Println("Error from get_all_topics")
			return nil, err
		}

		type ExtendedTopic struct {
			Topic  Topic
			Status string
		}

		var extendedTopics []ExtendedTopic

		for _, topic := range allTopics {
			var temp ExtendedTopic
			temp.Topic = topic
			temp.Status = "closed"

			expireTime, errTimeParse := time.Parse(time.RFC3339, topic.ExpireDate)
			if errTimeParse != nil {
				fmt.Println(errTimeParse)
				return nil, errTimeParse
			}
			if time.Now().Before(expireTime) {
				fmt.Println("before has user voted args")
				fmt.Println(topic.ID, account.Email)
				userVoted, err := t.hasUserVoted(stub, []string{topic.ID, account.Email})
				fmt.Println(userVoted)
				if err != nil {
					fmt.Println(err)
					return nil, err
				}
				if userVoted {
					temp.Status = "voted"
				} else {
					temp.Status = "open"
				}

				//if topic has not closed, results should be hidden from viewers, so results are cleared
				if temp.Status != "open" {
					//for i := range temp.Topic.Votes {
					//temp.Topic.Votes[i] = "0"
					//}
				}
			}

			fmt.Println("Appending extended topic", temp)
			extendedTopics = append(extendedTopics, temp)
		}

		//json.Marshal can only marshal JSON, not array of JSON, so we put array inside single JSON object to pass to server
		type JSONcapsule struct {
			AllTopics []ExtendedTopic
		}
		extendedTopicsJSON := JSONcapsule{
			AllTopics: extendedTopics,
		}

		extendedTopicsBytes, err1 := json.Marshal(&extendedTopicsJSON)
		if err1 != nil {
			fmt.Println("Error marshalling extendedTopics")
			return nil, err1
		}
		fmt.Println("All success, returning extendedTopics")
		return extendedTopicsBytes, nil

	case "get_topic":
		if len(args) != 2 {
			fmt.Println("Incorrect number of arguments. Expecting 2: topic ID and user ID")
			return nil, errors.New("Incorrect number of arguments. Expecting 2: topic ID and user ID")
		}

		account, errAccount := t.getAccount(stub, []string{args[1]})
		if errAccount != nil {
			fmt.Println("Error getting account:", errAccount)
			return nil, errAccount
		}

		topic, errTopic := t.getTopic(stub, args[0])
		if errTopic != nil {
			fmt.Println("Error getting topic: ", errTopic)
			return nil, errTopic
		}

		var status = "closed"

		expireTime, errTimeParse := time.Parse(time.RFC3339, topic.ExpireDate)
		if errTimeParse != nil {
			fmt.Println(errTimeParse)
			return nil, errTimeParse
		}
		if time.Now().Before(expireTime) {
			userVoted, err := t.hasUserVoted(stub, []string{topic.ID, account.Email})
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			if userVoted {
				status = "voted"
			} else {
				status = "open"
			}
		}

		type ExtendedTopic struct {
			Topic  Topic
			Status string
		}

		var extendedTopic ExtendedTopic
		extendedTopic.Topic = topic
		extendedTopic.Status = status

		//if topic has not closed, results should be hidden from viewers, so results are cleared
		if extendedTopic.Status != "open" {
			//for i := range extendedTopic.Topic.Votes {
			//extendedTopic.Topic.Votes[i] = "0"
			//}
		}

		topicBytes, errMarshal := json.Marshal(&extendedTopic)
		if errMarshal != nil {
			fmt.Println("Error marshalling extended topic: ", errMarshal)
			return nil, errMarshal
		}
		return topicBytes, nil

	case "get_account":
		if len(args) != 1 {
			fmt.Println("Incorrect number of arguments. Expecting 1: string of account ID being queried")
			return nil, nil
		}

		fmt.Println("[GET ACCOUNT ARGS]", args)

		accountID := args[0]
		account, err1 := t.getAccount(stub, []string{accountID})

		if err1 != nil {
			fmt.Println("Error from get_account: ", err1)
			return nil, err1
		}

		accountBytes, err2 := json.Marshal(&account)
		if err2 != nil {
			fmt.Println("Error marshalling account: ", err2)
			return nil, err2
		}
		fmt.Println("All success, returning account")
		return accountBytes, nil

	case "tally_votes":
		if len(args) != 1 {
			fmt.Println("Incorrect number of arguments. Expecting 1: string of topic ID being tallied")
			return nil, nil
		}

		topicID := string([]byte(args[0]))

		strArgs := []string{topicID}

		topicVotes, err1 := t.tallyVotes(stub, strArgs)
		if err1 != nil {
			fmt.Println("Error from tally_votes: ", err1)
			return nil, err1
		}

		topicVotesBytes, err2 := json.Marshal(&topicVotes)
		if err2 != nil {
			fmt.Println("Error marshalling vote tallies: ", err2)
			return nil, err2
		}
		fmt.Println("All success, returning vote tallies")
		return topicVotesBytes, nil

	case "get_open_requests":
		fmt.Println("I am in get open requests")
		allOpenRequests, err := t.getOpenRequests(stub)

		fmt.Println("All open Reqs:")
		fmt.Println(allOpenRequests)
		if err != nil {
			fmt.Println("Error from get_all_topics")
			return nil, err
		}
		//json.Marshal can only marshal JSON, not array of JSON, so we put array inside single JSON object to pass to server
		type JSONcapsule struct {
			AllAccReq []Account
		}
		AccReqJSON := JSONcapsule{
			AllAccReq: allOpenRequests,
		}
		fmt.Println("AccReqJSON")
		fmt.Println(AccReqJSON)
		allOpenRequestsBytes, err1 := json.Marshal(&AccReqJSON)

		fmt.Println("All open Reqs bytes:")
		fmt.Println(allOpenRequestsBytes)
		if err1 != nil {
			fmt.Println("Error marshalling allOpenRequests")
			return nil, err1
		}
		fmt.Println("All success, returning allOpenRequests")
		return allOpenRequestsBytes, nil
	}
	fmt.Println("query did not find func: " + function) //error

	return nil, errors.New("Received unknown function query")
}

// Init resets all the things
func (t *SimpleChaincode) Init(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	if len(args) > 0 {
		err := stub.PutState("InitState", []byte(args[0]))
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
	}

	fmt.Println("Initializing vote topics...")
	var blank []string
	blankBytes, _ := json.Marshal(&blank)
	err := stub.PutState("VoteTopics", blankBytes)
	if err != nil {
		fmt.Println("Failed to initialize vote topics")
	} else {
		fmt.Println("Successfully initialized vote topics")
	}

	fmt.Println("Initializing existing accounts...")
	errAcc := stub.PutState("accounts", blankBytes)
	if errAcc != nil {
		fmt.Println("Failed to initialize vote topics")
	} else {
		fmt.Println("Successfully initialized vote topics")
	}

	fmt.Println("Initializing cast votes...")
	blankBytes2, _ := json.Marshal(&blank)
	err2 := stub.PutState("CastVotes", blankBytes2)
	if err2 != nil {
		fmt.Println("Failed to initialize cast votes")
	} else {
		fmt.Println("Successfully initialized cast votes")
	}

	//create table to store all the user account requests
	errAccountRequest := stub.CreateTable("AccountRequests", []*shim.ColumnDefinition{
		&shim.ColumnDefinition{Name: "email", Type: shim.ColumnDefinition_STRING, Key: true},
		&shim.ColumnDefinition{Name: "full_name", Type: shim.ColumnDefinition_STRING, Key: false},
		&shim.ColumnDefinition{Name: "status", Type: shim.ColumnDefinition_STRING, Key: false},
		&shim.ColumnDefinition{Name: "org", Type: shim.ColumnDefinition_STRING, Key: false},
		&shim.ColumnDefinition{Name: "privileges", Type: shim.ColumnDefinition_STRING, Key: false}, //stored as an array in one string, with ' ' designating a new privilege (e.g. "manager creater")
		&shim.ColumnDefinition{Name: "time", Type: shim.ColumnDefinition_STRING, Key: false},
	})
	// Handle table creation errors
	if errAccountRequest != nil {
		fmt.Println(fmt.Sprintf("[ERROR] Could not create account request table: %s", errAccountRequest))
		return nil, errAccountRequest
	}

	//create table to store all the user account requests
	errApprovedAccount := stub.CreateTable("ApprovedAccounts", []*shim.ColumnDefinition{
		&shim.ColumnDefinition{Name: "userID", Type: shim.ColumnDefinition_STRING, Key: true},
		&shim.ColumnDefinition{Name: "full_name", Type: shim.ColumnDefinition_STRING, Key: false},
		&shim.ColumnDefinition{Name: "email", Type: shim.ColumnDefinition_STRING, Key: false},
		&shim.ColumnDefinition{Name: "org", Type: shim.ColumnDefinition_STRING, Key: false},
		&shim.ColumnDefinition{Name: "privileges", Type: shim.ColumnDefinition_STRING, Key: false}, //stored as an array in one string, with ' ' designating a new privilege (e.g. "manager creater")
		&shim.ColumnDefinition{Name: "votes", Type: shim.ColumnDefinition_UINT64, Key: false},
		&shim.ColumnDefinition{Name: "req_time", Type: shim.ColumnDefinition_STRING, Key: false},
		&shim.ColumnDefinition{Name: "appr_time", Type: shim.ColumnDefinition_STRING, Key: false},
		&shim.ColumnDefinition{Name: "appr_manager", Type: shim.ColumnDefinition_STRING, Key: false},
	})
	// Handle table creation errors
	if errApprovedAccount != nil {
		fmt.Println(fmt.Sprintf("[ERROR] Could not create account request table: %s", errApprovedAccount))
		return nil, errApprovedAccount
	}

	_, _ = t.requestAccount(stub, []string{"", "manager", "", "manager,creator"})
	_, _ = t.changeStatus(stub, []string{"approved", "", "manager", "", "manager,creator", "", "10"})
	return nil, nil
}
