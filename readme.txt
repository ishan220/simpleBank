#########
docker ps --to show the running containers, container is running instance of a image  
######
docker pull postgres:16-alpine
######
docker images
#### 
docker run --name container_name -p hostnetorkport:containernetwork port -e <env var> -d(detached/background) image
docker run --name postgres16 -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:16-alpine
#####
docker exec interactivemode container_name commandtotun -U <user>
docker exec -it postgres16 psql -U root
###
docker logs container_name
####
docker stop containername/containerId
###
docker run containerName/containerId
####
docker ps -a //all the running and stopped container
####
docker start containername/containerId
###
docker exec -it containername /bin/sh  --to run postgres in container shell
createdb --username=root --owner=root simple_bank
psql simple_bank
drop db simple_bank
docker exec -it postgres16 createdb --username=root --owner=root simple_bank
docker stop <containername>
docker rm <containername>
gmake postgres //execute the command mentioned in makefile
gmake createdb //execute the command mentioned in makefile
######
migrate -path=db/migration -database "postgres://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose up  //it executes the upmigration script
gmake migrateup
#######
\q  --> to quit psql prompt

#######
brew install golang-migrate
migrate -help
migrate create -ext sql -dir db/migration -seq init_schema //this command creates below 2 files
/Users/ishan/Project Folder/Projects_folders/Golang/Udemy_course/db/migration/000001_init_schema.up.sql
/Users/ishan/Project Folder/Projects_folders/Golang/Udemy_course/db/migration/000001_init_schema.down.sql
###### 
Queries framework/Options:
SQL  -- very fast and straightforward, Manual mapping sql fields to variables,
        Easy to make mistakes , not caught until runtime.
Gorm -- Crud Operations Already Implemented, very short production code,
        Must learn to write queries using gorm function.
Sqlx -- easy to use, Field mapping via query text & struct tags , 
        failure won't occur during runtime. 
Sqlc -- easy to use, works for postgres,  
        for MYSql is experimental, errors can be found before runtime

-->brew install kyleconroy/sqlc/sqlc
-->sqlc help --to get the diff commands
-->sqlc init  -->woukd create the yaml file
##########
Yaml file params
 - name: "db"          //name of the golang package that would be generated
    path: "./db/sqlc"  //path for the generated golang code
    queries: "./db/query/" where to look for sql query files
    schema: "./db/migration/"
    engine: "postgresql"   //engine used
    emit_json_tags: true   // to add json tags to generate stags
    emit_prepared_queries: true    // work with prepared statements
    emit_interface: false  // to generated interface , to mock the database and use for higher order details
    emit_exact_table_names: false // sync of struct names with table name



###########

--  name : CreateAccount :one        --->this denotes the name of the function signature in go , one denote the object returned by the function
INSERT INTO accounts (
    owner,
    balance,
    currency
) VALUES (
    $1,$2,$3
) RETURNING *            --> this will return all the inserted values

gmake sqlc  --this will generate the models.go , db.go, <sqlfile>.go file

####
-- name: UpdateAccount :one
UPDATE accounts 
SET balance = $2
WHERE id = $1
RETURNING *;
 or 
-- name: UpdateAccount :exec
UPDATE accounts 
SET balance = $2
WHERE id = $1;


########
Test file(main_test.go)
go get github.com/lib/pq
add in test file as below , adding _ before it will avoid it from removing from import if no function is used from that package
_ "github.com/lib/pq"


Test file(account_test.go)
go get github.com/stretchr/testify/require
######

In this example foreign key is causing the deadlocks 
as INSERT into Transfer is not allowing select * from account query , 
as there is a foreign key constraint.


################
reason for locks in DB transactions

1)If there is a table having foreign key to another primary table,
Suppose a record is inserted in this secondary table 
and we try to select same account from Primary Table, it can create a deadlock, since on
inserting a record in secondary table it references primary table.

like in this 
secondary table : Inserting a record in transfers table , and then select (get query) from accounts table
can create deadlock

for that in Select (get account query), the below mentioned NO KEY Update is added
SELECT * FROM accounts
WHERE id=$1 LIMIT 1
FOR NO KEY UPDATE; 

2) The order of queries change in two transactions

In Tx1, money is going from account1 and updated first
 and 

 In Tx2, money is coming in account2 first

Now in txn1
  secondly money is coming in account2 and updated, since tx1 has to update account2  , it will wait for tx2 to release the lock 

Now in tx2, account1 has to be updated , but its locked is held by tx1,
and tx1 is waiting  is waiting to release lock from tx2 on account2

so this creates a deadlock


#####
Isolation
To set isolation level in MYSql:
-->set session transaction isolation level serializable.

Dirty Read(Read Uncommitted) : Reading into another tranaction updation before its committed.

REad Committed(Non Repeatable read aka Phantom read) :  T2  will read after t1 has committed.

Repeatable read   :It will not read the updated made by other committed/uncommitted transactions,
but when updating the result in its own tranaction , it will consider the updation made by other transaction.

 Serializable: If one trnsaction is updating the table and second transaction is executing get query on same table,
 it will wait the other txn to commit or ele it will create timeout ,

 and if second trnsaction also tries to updated ,w hen first txn also is doing the same , it will
 create deadlock.

 In POSTGRES:

 1)Read uncommitted behave as read committed.
 2) In repeatable read scond txn will throw erro if we update ,saying first txn already updated.


###
git push steps

git init  --in root directory of Project
git add .
git commit -m "message"
git push --set-upstream https://github.com/ishan220/simpleBank.git main
enter username:ishan220
enter password:settings->developer settings->gen personal token->copy paste that token