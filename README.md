# Description
Playground est une application permettant de trouver le terrain de sport le plus proche de chez soi.
La partie back-end va se charger de stocker les terrains et de les envoyer au format JSON.

# Objectifs de ce projet :
- Améliorer mes compétences en :
    - Golang
    - Design d'application
    - TDD
    - Réaliser le déploiement d'un site en HTTPS avec un nom de domaine acheté
    - Mettre en oeuvre les connaissances apprises en NodeJS, docker, HTML et CSS.
    - Réaliser un site utile à terme :-)
# Architecture
## Front-end
Simple fichier HTML + JS
## Back-end
La base de données sera un fichier JSON dans un 1er temps puis PostgreSQL.

# Fonctionnalités (par ordre de priorité)
- Afficher l'ensemble des terrains de basket parisien
    - Back-end :
        1. Récupérer le path vers le fichier JSON  contenant les terrains et le parser
        1. Enregistrer les terrains dans une BDD
        1. Créer un webserver écoutant les requêtes Get à l'adresse "/terrains"
        1. Envoyer la liste des terrains triés par nom à travers une API sous format JSON
            > Faire attention aux HTTP status codes, Content-Type, méthodes HTTP utilisées, Accept-Encoding

    - Front-end :
        1. Mettre en place un static file server
        1. Mettre en place un layout général et un template spécifique à cette page
        1. Récupérer et Afficher la liste des terrains
- Afficher le détail d'un terrain
    - Back-end :
    > Pré-requis :
    Les terrains sont stockés dans une BDD

    1. Ecouter les requetes à "/api/playgrounds/{ID}"
    1. Récupérer l'ID du terrain depuis l'url de la requête
    1. Interroger la BDD pour retourner le terrain
    1. Envoyer le terrain à travers une API sous format JSON
        > Faire attention aux HTTP status codes, Content-Type, méthodes HTTP utilisées, Accept-Encoding
- Renvoyer les terrains par ordre de proximité pour une adresse donnée
    1. Récupérer l'adresse
    2. Récupérer la longitude et lattitude de cette adresse
    3. Calculer le delta par rapport à tous les playgrounds
    4. Retourner un tableau de terrains triés par ordre croissant de distance

- Mettre en place l'installation de la PWA
    1. Créer le manifest.json
    1. Enregister le Service worker dans main.js
    1. Ajouter le sw.js
    1. Ajouter la pop-up d'installation
- Se connecter en OpenID et envoyer un JWT Token
    1. Mettre en place l'OpenID avec les différents providers
    1. Envoyer la requête aux providers
    1. Récupérer les informations des providers
    1. Forger un JWT avec les informations utilisateur nickname ou firstname ou email
        > Utiliser Goth pour homogénéiser la gestion des providers
        > Utiliser github.com/dgrijalva/jwt-go pour mettre en place le JWT
        > Mettre les ID client et secret key en tant que variables d'environnement
- Mettre à jour l'UI en fonction de la présence d'un JWT
- Déconnexion du compte
    1. Invalider le JWT
- Mettre en place le refresh middleware
    1. Parser le cookie
    1. Si le cookie est valide, réinitialiser le temps d'expiration
- Commenter les terrains en étant connecté
- Ajouter de nouveaux terrains en étant connecté
    > Si le terrain existe déjà renvoyer un statut badrequest
    > Vérifier la casse des noms de terrains au moment de les ajouter
- Intégrer Googlemap
- Disposer d'une fonction de recherche en fonction de certains critères (arrondissement, nom, horaires d'ouverture)
- En disposant d'un profil modérateur ou administrateur, accepter les demandes d'ajout de nouveaux terrains
    1. Parser le cookie
    1. Si le cookie n'est pas valide, rediriger vers la page de login
        > Ajouter les rôles dans le JWT
        > Ajouter un middleware d'autorisation d'accès à certaines pages
- Afficher l'itinéraire (à pied, en voiture, le meilleur transport)
- Pouvoir noter les terrains en étant connecté
- Mettre en place des évènements et un calendrier pour chaque terrain (pour que des joueurs puisse convenir sur un horaire de RDV)
- Auto-complétion de l'adresse
- Créer un compte
    1. Récupérer et parser le contenu de la POST request
    2. Créer une entrée dans la BDD
    3. Retourner le status Accepted
- Se connecter à son compte et recevoir un JWT Token
- Modifier son profil
- Ajouter des photos des terrains
- Ajouter l'utilisation du cache pour les static assets et les appels d'API avec la PWA
    > https://www.julienpradet.fr/fiches-techniques/pwa-intercepter-les-requetes-http-et-les-mettre-en-cache/
    > https://www.julienpradet.fr/fiches-techniques/pwa-declarer-un-service-worker-et-gerer-son-cycle-de-vie/
-Publier l'application sur le PlayStore
    > Utiliser https://appmaker.xyz/pwa-to-apk/ pour publier l'application sur le PlayStore

# Back-end

Le webscraper produit un fichier JSON qui sert ensuite à peupler la base de données de l'application.
Pour cela, les données ont été récupérées depuis plusieurs sources (Open data, web scraping).
Il est possible de créer un compte permettant de commenter les terrains et d'en soumettre de nouveaux.

# TODO

- Réaliser le front-end
- Ajouter une fonction permettant de lancer le serveur en https en mode production et sur le port 8080 de localhost en mode dev.
- Rediriger le traffic HTTP vers HTTPS
- Automatiser le renouvellement du certificat TLS.

### Web Scraper (NodeJS, librairie : puppeteer/cheerio)

- ~~Data scraper~~

### Application :

- Liste des terrains parisiens
- Notation des terrains
- Commentaires
- Responsive
- Niveau de jeu des terrains
- Localisation des terrains (Gmap)
- Description des terrains
- Page profil de l'utilisateur
- Utilisation de JWT pour garder la session active
- Utilisation de PostgreSQL pour enregistrer les utilisateurs, terrains et commentaires
- Mise en place de filtres
- PWA
- Soumettre de nouveaux terrains et créer une page admin pour les accepter
- Photos
- Réconciliation de données
- Recherche par arrondissement
- Agenda des terrains et création de communautés
- Déployer l'application en https

# Améliorations
- Ajouter de la concurrence
- Mettre en place un cache
- Ajouter d'autres jeux de données : https://data.iledefrance.fr/explore/dataset/20170419_res_fichesequipementsactivites/information/?disjunctive.actlib

## Sources de données

- https://www.gralon.net/mairies-france/paris/equipements-sportifs-terrain-de-basket-ball-75056.htm
- http://www.cartes-2-france.com/activites/750560006/ritz-health-club.php donne accès aux liens https://www.webvilles.net/sports/75056-paris.php
- https://www.data.gouv.fr/fr/datasets/recensement-des-equipements-sportifs-espaces-et-sites-de-pratiques/
