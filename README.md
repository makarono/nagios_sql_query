# nagios_sql_query
run mysql and oracle query in nagios and icinga

# get usage:

./query_db.go -help


# examples:

./query_db.go -host="10.10.11.128" -user="xxxxxx" -password="xxxxxx" -query="select 22 from dual" -shema="ppc_lt" -timeout=1s

./query_db.go -host="10.10.11.128" -user="xxxxxx" -password="xxxxxx" -query="select count(1) +18 from dual" -shema="ppc_lt" -timeout=10000000s -warning="19" -critical=30 -dbtype=mysql

./query_db.go -host="10.10.10.151" -user="xxxx" -password="xxxx" -servicename="foo" -port=1521 -query="select 1 from dual" -warning="19" -critical=30 -dbtype=oracle

./query_db.go -host="10.10.10.128" -user="xxxxxx" -password="xxxxxx" -query="select 0 from dual" -shema="ppc_lt" -timeout=10000000s  -critical=0 -dbtype=mysql -inverse -message=test message
