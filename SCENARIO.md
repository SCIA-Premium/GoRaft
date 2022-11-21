# Without errors

LOAD toto \
APPEND toto "This is toto" \

---

LOAD toto \
DELETE toto \

---

LOAD toto \
LOAD tutu \

---

SPEED *follower* low
LOAD toto \

=> The *follower* should only register one log entry

# With errors

LOAD toto \
LOAD toto \

=> Double load

---

DELETE toto \

=> Delete without load

---

LOAD toto \
CRASH *leader* \

---

CRASH *follower* \
LOAD toto \
RECOVERY *follower* \

---

LOAD toto \
CRASH *leader* (**before** commiting) \
RECOVERY *leader* \

=> The load should **not** be done

---

LOAD toto \
CRASH *leader* (**after** commiting but **before** validating) \
RECOVERY *leader* \

=> The load should be done

---

CRASH *follower* \
CRASH *leader* \
RECOVERY *leader* \
LOAD toto \
LOAD tata \
RECOVERY *follower* \