package parser

import (
	"reflect"
	"testing"
)

func TestExtractTables(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		want    []string
		wantErr bool
	}{
		{
			"simple select query",
			"select * from cities",
			[]string{"cities"},
			false,
		},
		{
			"no table select query",
			"SELECT 5 * 3 AS result;",
			[]string{},
			false,
		},
		{
			"select query with alias",
			"select * from country c",
			[]string{"country"},
			false,
		},
		{
			"select query with schema",
			"select * from public.country c",
			[]string{"country"},
			false,
		},
		{
			"select query with join",
			"select * from city c inner join country c2 on c.countrycode = c2.code where id < 100 limit 90",
			[]string{"city", "country"},
			false,
		},
		{
			"select from select",
			`select n1.name, n1.author_id, count_1, total_count
from (select id, name, author_id, count(1) as count_1
      from names
      group by id, name, author_id) n1
inner join (select id, author_id, count(1) as total_count
          from names
          group by id, author_id) n2
on (n2.id = n1.id and n2.author_id = n1.author_id)`,
			[]string{"names"},
			false,
		},
		{
			"select from select from select",
			`SELECT ens.company, ens.state, ens.zip_code, ens.complaint_count
FROM (select company, state, zip_code, count(complaint_id) AS complaint_count
   FROM credit_card_complaints
   WHERE state IS NOT NULL
   GROUP BY company, state, zip_code) ens
INNER JOIN
(SELECT ppx.company, max(ppx.complaint_count) AS complaint_count
 FROM (SELECT ppt.company, ppt.state, max(ppt.complaint_count) AS complaint_count
       FROM (SELECT company, state, zip_code, count(complaint_id) AS complaint_count
             FROM credit_card_complaints_2
             WHERE company = 'Citibank'
              AND state IS NOT NULL
             GROUP BY company, state, zip_code
             ORDER BY 4 DESC) ppt
       GROUP BY ppt.company, ppt.state
       ORDER BY 3 DESC) ppx
 GROUP BY ppx.company) apx
ON apx.company = ens.company
AND apx.complaint_count = ens.complaint_count
ORDER BY 4 DESC;`,
			[]string{"credit_card_complaints", "credit_card_complaints_2"},
			false,
		},
		{
			"select from function",
			`SELECT setup::date
FROM generate_series('2007-02-01', '2007-02-28', INTERVAL '1 day') AS setup`,
			[]string{},
			false,
		},
		{
			"select from tables",
			`SELECT date_part('day', p.payment_date)::INT AS legit,
SUM(p.amount),
date_part('day', fk.setup)::INT AS fake
FROM payment AS p
LEFT JOIN fake_month AS fk
ON date_part('day', fk.setup)::INT = date_part('day', p.payment_date)::INT
GROUP BY legit, fake
HAVING SUM(p.amount) > $1
ORDER BY fake NULLS first
LIMIT $2;`,
			[]string{"payment", "fake_month"},
			false,
		},
		{
			"select using with clause",
			`WITH fake_month AS(
SELECT setup::date
FROM generate_series('2007-02-01', '2007-02-28', INTERVAL '1 day') AS setup
)
SELECT date_part('day', p.payment_date)::INT AS legit,
SUM(p.amount),
date_part('day', fk.setup)::INT AS fake
FROM payment AS p
LEFT JOIN fake_month AS fk
ON date_part('day', fk.setup)::INT = date_part('day', p.payment_date)::INT
GROUP BY legit, fake
HAVING SUM(p.amount) > $1
LIMIT $2;`,
			[]string{"payment"},
			false,
		},
		{
			"select in join part",
			`SELECT pg_database.datname,tmp.mode,COALESCE(count,$1) as count
FROM ( VALUES ($2), ($3), ($4), ($5), ($6), ($7), ($8), ($9) ) AS tmp(mode)
CROSS JOIN pg_database
LEFT JOIN (SELECT database, lower(mode) AS mode,count(*) AS count FROM pg_locks WHERE database IS NOT NULL GROUP BY database, lower(mode) ) AS tmp2
ON tmp.mode=tmp2.mode and pg_database.oid = tmp2.database ORDER BY 1`,
			[]string{"pg_database", "pg_locks"},
			false,
		},
		{
			"select with intersect",
			`SELECT count(*) FROM (SELECT * FROM without_complaints
    INTERSECT
    SELECT * FROM credit_card_wo_complaints) ppg`,
			[]string{"without_complaints", "credit_card_wo_complaints"},
			false,
		},
		{
			"select with except",
			`SELECT count(*) FROM (SELECT * FROM without_complaints
    EXCEPT
SELECT * FROM credit_card_wo_complaints) ppg`,
			[]string{"without_complaints", "credit_card_wo_complaints"},
			false,
		},
		{
			"select in where clause",
			`select d.*
from order_line_detail d
where d.line_id in (
  select l.id
  from order_line l
  where l.order_id in (
      select o.id
      from orders o
      where o.last_update > now() - interval '6 hours'
  )
);`,
			[]string{"order_line_detail", "order_line", "orders"},
			false,
		},
		{
			"select in where clause 2",
			`select * from city c inner join country c2 on c.countrycode = c2.code where countrycode = (SELECT c3.countrycode from countrylanguage c3 where c3.countrycode = 'KGZ' limit 1) limit 90`,
			[]string{"city", "country", "countrylanguage"},
			false,
		},
		{
			"select from select in where clause",
			`SELECT first_name, last_name, email
FROM customer
WHERE customer_id IN (
SELECT customer_id FROM (
SELECT DISTINCT customer_id, SUM(amount)
FROM payment
WHERE extract(month from payment_date) = 4
AND extract(day from payment_date) BETWEEN 10 AND 13
GROUP BY customer_id
HAVING SUM(amount) > 30
ORDER BY SUM(amount) DESC
LIMIT 5) AS top_five);`,
			[]string{"customer", "payment"},
			false,
		},
		{
			"select in where clause",
			`with cte_order as (
    select o.id 
    from orders o
    where o.last_update > now() - interval '6 hours'
),
cte_line as (
    select l.id
    from order_line l
    where l.order_id in (
        select * from cte_order
    )
)
select d.*
from order_line_details d
where d.id in (select * from cte_line)`,
			[]string{"order_line_details", "orders", "order_line"},
			false,
		},
		{
			"select using with and union",
			`WITH RECURSIVE search_graph(id, link, data, depth, path, cycle) AS (
        SELECT g.id, g.link, g.data, 1,
          ARRAY[g.id],
          false
        FROM graph g
      UNION ALL
        SELECT g.id, g.link, g.data, sg.depth + 1,
          path || g.id,
          g.id = ANY(path)
        FROM graph_2 g, search_graph sg
        WHERE g.id = sg.link AND NOT cycle
)
SELECT * FROM search_graph;`,
			[]string{"graph", "graph_2"},
			false,
		},
		{
			"select with union",
			`SELECT
   column_1,
   column_2
FROM
   tbl_name_1
UNION
SELECT
   column_1,
   column_2
FROM
   tbl_name_2;
`,
			[]string{"tbl_name_1", "tbl_name_2"},
			false,
		},
		{
			"wrong select query",
			`SELECT count(*) FROM (SELECT * FROM without_complaints
     EXCEPT
SELECT * FROM credit_card_wo_complaints)`,
			nil,
			true,
		},

		// INSERT queries
		{
			"simple insert query",
			"insert into city (name, countrycode, district, population) values ('bishkek', 'kgz', 'bishkek', 1000000)",
			[]string{"city"},
			false,
		},
		{
			"insert query with select",
			`INSERT INTO sales.big_orders (id, full_name, address, total)
SELECT
   id,
   full_name,
   address,
   total
FROM
   sales.total_orders
WHERE
   total > $1;
`,
			[]string{"big_orders", "total_orders"},
			false,
		},

		// UPDATE queries
		{
			"simple update query",
			`update cities SET a=b;`,
			[]string{"cities"},
			false,
		},
		{
			"update query with select",
			` UPDATE employees SET sales_count = sales_count + 1 WHERE id =
   (SELECT sales_person FROM accounts WHERE name = 'Acme Corporation');`,
			[]string{"employees", "accounts"},
			false,
		},
		{
			"update query with when case",
			`UPDATE reward_members
SET member_status = (
CASE member_status
WHEN 'gold' THEN 'gold_group'
WHEN 'bronze' THEN 'bronze_group'
WHEN 'platinum' THEN 'platinum_group'
WHEN 'silver' THEN 'silver_group'
END
)
WHERE member_status IN ('gold', 'bronze','platinum', 'silver');`,
			[]string{"reward_members"},
			false,
		},

		// DELETE queries
		{
			"simple delete query",
			`DELETE FROM tbl_scores`,
			[]string{"tbl_scores"},
			false,
		},
		{
			"delete query with select in where clause",
			`DELETE FROM tbl_scores
WHERE student_id IN
(SELECT student_id
FROM
(SELECT student_id,
ROW_NUMBER() OVER(PARTITION BY student_id
ORDER BY student_id) AS row_num
FROM tbl_scores_2) t
WHERE t.row_num <> 1);
`,
			[]string{"tbl_scores", "tbl_scores_2"},
			false,
		},

		// Complex queries
		{
			"complex query 1",
			`WITH upd AS (
UPDATE employees SET sales_count = sales_count + 1 WHERE id =
  (SELECT sales_person FROM accounts WHERE name = 'Acme Corporation')
  RETURNING *
)
INSERT INTO employees_log SELECT *, current_timestamp FROM upd;
`,
			[]string{"employees_log", "employees", "accounts"},
			false,
		},
		{
			"complex select query",
			`with order_price_transport as (
    select orders.id, price_parts.code, price_parts.status, sum(price_parts.amount)
    from orders
    join price_parts on price_parts.order_id = orders.id
    where price_parts.code in ('FUEL', FREIGHT)
    group by orders.id, price_parts.code, price_parts.status
),
order_price_other as (
    select orders.id, price_parts.code, price_parts.status, sum(price_parts.amount)
    from orders
    join price_parts on price_parts.order_id = orders.id
    where price_parts.code not in ('FUEL', 'FREIGHT')
    group by orders.id, price_parts.code, price_parts.status
)
select
    orders.id,
    est_trans_price.amount,
    act_trans_price.amount,
    est_other_price.amount,
    act_other_price.amount
from orders
join order_price_transport as est_trans_price on est_trans_price.id = orders.id and est_trans_price.status = 'estimated'
join order_price_transport as act_trans_price on act_trans_price.id = orders.id and act_trans_price.status = 'actual'
join order_price_other as est_other_price on est_other_price.id = orders.id and est_other_price.status = 'estimated'
join order_price_other as act_other_price on act_trans_price.id = orders.id and act_other_price.status = 'actual'
where orders.creation_date between $1 and $2
and orders.organization_id in ($3)
and orders.status_id in (123, 456)
`,
			[]string{"orders", "price_parts"},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Parser{}
			got, err := p.ExtractTables(tt.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parser.ExtractTables() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parser.ExtractTables() = %v, want %v", got, tt.want)
			}
		})
	}
}
