# Relay Pivot

> **Status:** source brief. The product definition is in [`product.md`](product.md);
> the ordered implementation plan is in [`pivot-plan.md`](pivot-plan.md). This
> file is kept as the original note that started the pivot.

Previously relay had a pretty contrarian / marxist view of the world. it was created basically as an attack against cloud providers.
This view continues but we believe now that is most important to provide an alternative than just attack and ship an ideology with a software product.

we believe that in order to make relay truly usefull and a clear alternative to platforms and at the same time build a better world we need to pivot quite harshly.

## New vision:
Relay is a virtual third space - a place for you and for your friends to share photos, images, ideas, articles, links and your own space in the web. Hosted by your friends,
it is as easy and straightforward as installing a minecraft server. You have the server and your friends - and only your friends, or friends of your friends (see our invitation mode) can post on it.

It's a micro social network where everyone has a main page with a list of all the websites and stuff that they share alongside a main page where you can see everyones.
You can comment on your friends posts. - no retweet, as a relay node is just a node of about 4 to 29 friends so you will continue to see everyone's posts easily-

## Invitation mechanism:
whoever hosts the node can create invitation links, the invitation link has a key and that is your friend - or friendgroup - account. people that has a key can also share a subkey via invitation links so the community can grow. No accounts. Only invitation links.

## No baked in ideology
the ideology that inspired relay doesnt need to be bundled in relay. we create a virtual third space for your real friends to connect and that's it. no need to subscribe to an anti-system vision of the world, just use it. we believe that pivoting from platform owned social networks is already and by itself a good change on the world. no need to say it.

## Microwebsites as templates for posting
Relay at heart continues being what it is, a system that hosts static websites. But in order to make the life of our non technical guys a bit easier when creating a new post a set of premade static websites named templates will be shown. For example if i want to publish an article a medium-like template will be available to be choose and prefilled on the preview page of the publish. i can edit fields and do changes on the template via a "prepublish.html" file that the template will have and then on publish the html file will call relay with a publish.js and its contents that will generate the real index.html with the template + the content by the user. The logic of prefilling and generating fields are on the template side. Relay will only run prepublish.html and a new endpoint named publish that the html will call with the user content algonside the publish.js that will host the code. Only valid keys may call that publish endpoint. Templates will be able to be added as normal websites inside relay.

## No accounts.
invitation links already contain the user key. it should be treated as a logged in user and show relay welcome screen, that will contain a button to go to the publish screen and the new posts of your friends alongside the profile of all of them. Clicking on a profile shows that friend page with all the content that he shared. clicking on a content should open the template or website and you will have a place to comment on their posts.
