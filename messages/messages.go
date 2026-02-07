// This file is part of club-1/newsletter-go.
//
// Copyright (c) 2026 CLUB1 Members <contact@club1.fr>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package messages

var (
	AlreadySubscribed_subject = Message{
		en: "Already subscribed",
		fr: "Déjà inscrit",
	}
	AlreadySubscribed_body = Message{
		en: "Your email is already subscribed, if problem persist, contact <%s>.",
		fr: "Votre email est déjà inscrit, si le problème persiste, contactez <%s>.",
	}
	ConfirmSubscription_subject = Message{
		en: "Please confirm your subsciption",
		fr: "Veuillez confirmer votre inscription",
	}
	ConfirmSubscription_body = Message{
		en: "Reply to this email to confirm that you want to subscribe to the newsletter [%s] (the content does not matter).",
		fr: "Répondez à cet email pour confirmez que vous souhaitez vous inscrire à la newsletter [%s] (le contenu n'a pas d'importance).",
	}
	ConfirmSubscriptionAlt_body = Message{
		en: "Reply to this email to confirm that you want to subscribe to %s's newsletter (the content does not matter).",
		fr: "Répondez à cet email pour confirmez que vous souhaitez vous inscrire à la newsletter de %s (le contenu n'a pas d'importance).",
	}
	SuccessfullSubscription_subject = Message{
		en: "Subscription is successfull !",
		fr: "Inscription réussie !",
	}
	SuccessfullSubscription_body = Message{
		en: "Your email has been successfully subscribed to the newsletter [%s].",
		fr: "Votre email a bien été inscrit à la newsletter [%s].",
	}
	SuccessfullSubscriptionAlt_body = Message{
		en: "Your email has been successfully subscribed to %s's newsletter.",
		fr: "Votre email a bien été inscrit à la newsletter de %s.",
	}
	SuccessfullUnsubscription_subject = Message{
		en: "Unsubscription is successfull",
		fr: "Désinscription réussie",
	}
	SuccessfullUnsubscription_body = Message{
		en: "Your email has been successfully unsubscribed from the newsletter [%s].",
		fr: "Votre email a bien été désinscrit de la newsletter [%s].",
	}
	SuccessfullUnsubscriptionAlt_body = Message{
		en: "Your email has been successfully unsubscribed from %s's newsletter.",
		fr: "Votre email a bien été désinscrit de la newsletter de %s.",
	}
	UnsubscriptionFailed_subject = Message{
		en: "Unsubscription failed",
		fr: "Échec de la désinscription",
	}
	UnsubscriptionFailed_body = Message{
		en: "Failed to unsubscribe your email from newsletter [%s]. Contact list owner for more infos: <%s>.",
		fr: "La désinscription de votre email à la newsletter [%s] a échoué. Contactez le propriétaire de la liste pour plus d'infos : <%s>",
	}
	UnsubscriptionFailedAlt_body = Message{
		en: "Failed to unsubscribe your email from %s's newsletter. Contact list owner for more infos: <%s>.",
		fr: "La désinscription de votre email à la newsletter de %s a échoué. Contactez le propriétaire de la liste pour plus d'infos : <%s>",
	}
	VerificationFailed_subject = Message{
		en: "Verification failed",
		fr: "Échec de la vérification",
	}
	VerificationFailed_body = Message{
		en: "Your email cannot be added to the subscripted list, contact list owner for more infos: <%s>.",
		fr: "Votre email ne peut pas être inscrit à la liste, veuillez contacter le propriétaire de la liste pour plus d'infos : <%s>.",
	}
)
