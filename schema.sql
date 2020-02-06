/*
 *    Copyright 2020 bitfly gmbh
 *
 *    Licensed under the Apache License, Version 2.0 (the "License");
 *    you may not use this file except in compliance with the License.
 *    You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *    Unless required by applicable law or agreed to in writing, software
 *    distributed under the License is distributed on an "AS IS" BASIS,
 *    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *    See the License for the specific language governing permissions and
 *    limitations under the License.
 */

drop table blocks;
drop table snarkjobs;
drop table feetransfers;
drop table userjobs;
drop table accounts;
drop table accounttransactions;
drop table daemonstatus;
drop table statistics;

create table if not exists blocks
(
    statehash         varchar(200) not null,
    canonical         bool         not null,
    previousstatehash varchar(200) not null,
    snarkedledgerhash varchar(200) not null,
    stagedledgerhash  varchar(200) not null,
    coinbase          int          not null,
    creator           varchar(200) not null,
    slot              int          not null,
    height            int          not null,
    epoch             int          not null,
    ts                timestamp    not null,
    totalcurrency     int          not null,
    usercommandscount int          not null,
    snarkjobscount    int          not null,
    feetransfercount  int          not null,
    primary key (statehash)
);
create index idx_blocks_creator on blocks (creator);
create index idx_blocks_ts on blocks (ts);
create index idx_blocks_height on blocks (height);

create table if not exists snarkjobs
(
    blockstatehash varchar(200) not null,
    index          int          not null,
    jobids         int[]        not null,
    prover         varchar(200) not null,
    fee            int          not null,
    primary key (blockstatehash, index)
);
create index idx_snarkjobs_prover on snarkjobs (prover);

create table if not exists feetransfers
(
    blockstatehash varchar(200) not null,
    index          int          not null,
    recipient      varchar(200) not null,
    fee            int          not null,
    primary key (blockstatehash, index)
);
create index idx_feetransfers_recipient on feetransfers (recipient);

create table if not exists userjobs
(
    blockstatehash varchar(200)  not null,
    index          int           not null,
    id             varchar(1000) not null,
    sender         varchar(200)  not null,
    recipient      varchar(200)  not null,
    memo           varchar(200)  not null,
    fee            int           not null,
    amount         int           not null,
    nonce          varchar(200)  not null,
    delegation     bool          not null,
    primary key (blockstatehash, index)
);

create table if not exists accounts
(
    publickey        varchar(200) not null primary key,
    balance          int          not null,
    nonce            int          not null,
    receiptchainhash varchar(200) not null,
    delegate         varchar(200) not null,
    votingfor        varchar(200) not null,
    txsent           int          not null,
    txreceived       int          not null,
    blocksproposed   int          not null,
    snarkjobs        int          not null,
    firstseen        timestamp    not null,
    lastseen         timestamp    not null
);
create index idx_accounts_firstseen on accounts (firstseen);

create table if not exists accounttransactions
(
    publickey varchar(200)  not null,
    id        varchar(1000) not null,
    ts        timestamp     not null,
    primary key (publickey, ts, id)
);

create table if not exists daemonstatus
(
    ts                         timestamp    not null primary key,
    blockchainlength           int          not null,
    commitid                   varchar(200) not null,
    epochduration              int          not null,
    slotduration               int          not null,
    slotsperepoch              int          not null,
    consensusmechanism         varchar(200) not null,
    highestblocklengthreceived int          not null,
    ledgermerkleroot           varchar(200) not null,
    numaccounts                int          not null,
    peers                      text[]       not null,
    peerscount                 int          not null,
    statehash                  varchar(200) not null,
    syncstatus                 varchar(200) not null,
    uptime                     int          not null
);

create table if not exists statistics
(
    indicator varchar(50) not null,
    ts        timestamp   not null,
    value     numeric     not null,
    primary key (indicator, ts)
);