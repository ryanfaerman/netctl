-- name: GetAccountKindMemberships :many
SELECT 
  a.*,
  r.id as role_id,
  r.permissions as role_permissions,
  r.ranking as role_ranking,
  r.name as role_name,
  m.created_at as membership_created_at
FROM accounts as a
JOIN memberships AS m ON a.id = m.member_of
JOIN roles AS r on m.role_id = r.id
WHERE m.account_id = ?1 AND a.kind = ?2;

-- name: GetMembershipsForAccount :many
SELECT memberships.*
FROM memberships
WHERE account_id = ?1;

-- name: GetMembershipsForAccountAndKind :many
SELECT memberships.*
FROM memberships
JOIN accounts ON memberships.member_of = accounts.id
WHERE memberships.account_id = ?1 AND accounts.kind = ?2;

-- name: CreateRoleOnAccount :one
INSERT INTO roles (
  name, account_id, permissions, ranking
) VALUES (?1, ?2, ?3, ?4)
RETURNING id;

-- name: GetRole :one
SELECT * FROM roles WHERE id = ?1;

-- name: CreateMembership :exec
INSERT INTO memberships (
  account_id, member_of, role_id, created_at
) VALUES (?1, ?2, ?3, CURRENT_TIMESTAMP);

-- name: FindMembershipForAccountMember :one
SELECT 
  memberships.*,
  roles.name as role_name,
  roles.permissions as permissions
FROM memberships 
Join roles on memberships.role_id = roles.id
WHERE memberships.account_id = ?1 AND member_of = ?2;

-- name: HasPermissionOnAccount :one
select memberships.* 
from memberships 
join roles on memberships.role_id = roles.id  
where 
  memberships.account_id=@account_id
  and member_of=@member_of
  and (roles.permissions & @permission) > 0;
