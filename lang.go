package newsletter

type Message struct {
	en string
	fr string
}

func (m Message) Print() string {
	switch Locale {
	case "en":
		return m.en
	case "fr":
		return m.fr
	}
	panic("internal error")
}

var Messages = struct {
	AlreadySubscribed_subject         Message
	AlreadySubscribed_body            Message
	ConfirmSubscription_subject       Message
	ConfirmSubscription_body          Message
	SuccessfullSubscription_subject   Message
	SuccessfullSubscription_body      Message
	SuccessfullUnsubscription_subject Message
	SuccessfullUnsubscription_body    Message
	VerificationFailed_subject        Message
	VerificationFailed_body           Message
}{
	AlreadySubscribed_subject: Message{
		en: "already subscribed",
		fr: "déjà inscri",
	},
	AlreadySubscribed_body: Message{
		en: "your email is already subscribed, if problem persist, contact <%s>",
		fr: "votre email est déjà inscris, si le problème persiste, contactez <%s>",
	},
	ConfirmSubscription_subject: Message{
		en: "confirm your subsciption",
		fr: "veuillez confirmer votre inscription",
	},
	ConfirmSubscription_body: Message{
		en: "Reply to this email to confirm that you want to subscribe to ",
		fr: "Répondez à cet email pour confirmez que vous souhaitez vous inscrire à ",
	},
	SuccessfullSubscription_subject: Message{
		en: "Subscription is successfull !",
		fr: "Inscription réussie !",
	},
	SuccessfullSubscription_body: Message{
		en: "Your email has been added to list ",
		fr: "Votre email a été ajouté à la liste ",
	},
	SuccessfullUnsubscription_subject: Message{
		en: "Unsubscription is successfull",
		fr: "Désinscription réussie",
	},
	SuccessfullUnsubscription_body: Message{
		en: "Your email was successfully removed from the list ",
		fr: "Votre email a bien été retiré de la liste ",
	},
	VerificationFailed_subject: Message{
		en: "Verification failed",
		fr: "Échec de la vérification",
	},
	VerificationFailed_body: Message{
		en: "Your email cannot be added to the subscripted list, contact list owner for more infos",
		fr: "Votre email ne peut pas être inscris à la liste, veuillez contacter le propriétaire de la liste pour plus d'infos",
	},
}
