# Overall Ethereum client interaction stucture

## P2P 

"Gossip" maintient les connections avec d'autres bloc et les propage, protocole eth/68 ou plus recent.
Ton client reçoit un message contenant le bloc encodé en RLP (Recursive Length Prefix).
Vérification de la taille du bloc 

## API Engine 

Le consensus reçoit le même bloc par son propre réseau et demande au client si le format est conforme au regles EVM
Prevalidation Header Check
Geth vérifie le parentHash (est-ce qu'il s'emboîte bien dans le bloc précédent ?), le timestamp, et le gasLimit.

## Exécution 

La tâche : Geth ouvre le "State" actuel (les soldes de tout le monde) et fait passer les transactions une par une dans l'EVM.

Le processus : 1. Pour chaque transaction, on vérifie la signature (ECDSA).
2. On déduit le Gaz.
3. On exécute le code du Smart Contract.
4. On met à jour les soldes et les données (le "World State").

Package clé : ```core/vm``` et ```core/state```.

## La Validation du Résultat (State Root)


La tâche : Une fois toutes les transactions exécutées, Geth calcule une nouvelle "empreinte digitale" de toute la base de données : le State Root (un arbre de Merkle-Patricia).
Comparaison : Geth compare son résultat avec le stateRoot écrit par le validateur dans le Header du bloc.
Verdict : Si les deux correspondent, Geth répond au Consensus : VALID. Sinon, INVALID.
La tâche : Si le bloc est valide et que le Consensus confirme qu'il devient "canonique" (appel engine_forkchoiceUpdated), Geth écrit les données sur le SSD.
Technique : On n'écrit pas juste "le bloc". On écrit les transactions, les reçus (receipts), et on met à jour le Trie (la structure de données de l'état).
Moteur : Geth utilise Pebble ou LevelDB (package ethdb).
