package newsletter

const (
	LangEnglish = "en"
	LangFrench  = "fr"
)

type Message struct {
	en string
	fr string
}

func (m Message) Print() string {
	switch Conf.Settings.Language {
	case LangEnglish:
		return m.en
	case LangFrench:
		return m.fr
	default:
		return m.en
	}
}

var Messages = struct {
	AlreadySubscribed_subject         Message
	AlreadySubscribed_body            Message
	ConfirmSubscription_subject       Message
	ConfirmSubscription_body          Message
	ConfirmSubscriptionAlt_body       Message
	SuccessfullSubscription_subject   Message
	SuccessfullSubscription_body      Message
	SuccessfullSubscriptionAlt_body   Message
	SuccessfullUnsubscription_subject Message
	SuccessfullUnsubscription_body    Message
	SuccessfullUnsubscriptionAlt_body Message
	VerificationFailed_subject        Message
	VerificationFailed_body           Message
}{
	AlreadySubscribed_subject: Message{
		en: "Already subscribed",
		fr: "Déjà inscrit",
	},
	AlreadySubscribed_body: Message{
		en: "Your email is already subscribed, if problem persist, contact <%s>.",
		fr: "Votre email est déjà inscrit, si le problème persiste, contactez <%s>.",
	},
	ConfirmSubscription_subject: Message{
		en: "Please confirm your subsciption",
		fr: "Veuillez confirmer votre inscription",
	},
	ConfirmSubscription_body: Message{
		en: "Reply to this email to confirm that you want to subscribe to the newsletter [%s] (the content does not matter).",
		fr: "Répondez à cet email pour confirmez que vous souhaitez vous inscrire à la newsletter [%s] (le contenu n'a pas d'importance).",
	},
	ConfirmSubscriptionAlt_body: Message{
		en: "Reply to this email to confirm that you want to subscribe to %s's newsletter (the content does not matter).",
		fr: "Répondez à cet email pour confirmez que vous souhaitez vous inscrire à la newsletter de %s (le contenu n'a pas d'importance).",
	},
	SuccessfullSubscription_subject: Message{
		en: "Subscription is successfull !",
		fr: "Inscription réussie !",
	},
	SuccessfullSubscription_body: Message{
		en: "Your email has been successfully subscribed to the newsletter [%s].",
		fr: "Votre email a bien été inscrit à la newsletter [%s].",
	},
	SuccessfullSubscriptionAlt_body: Message{
		en: "Your email has been successfully subscribed to %s's newsletter.",
		fr: "Votre email a bien été inscrit à la newsletter de %s.",
	},
	SuccessfullUnsubscription_subject: Message{
		en: "Unsubscription is successfull",
		fr: "Désinscription réussie",
	},
	SuccessfullUnsubscription_body: Message{
		en: "Your email has been successfully unsubscribed from the newsletter [%s].",
		fr: "Votre email a bien été désinscrit de la newsletter [%s].",
	},
	SuccessfullUnsubscriptionAlt_body: Message{
		en: "Your email has been successfully unsubscribed from %s's newsletter.",
		fr: "Votre email a bien été désinscrit de la newsletter de %s.",
	},
	VerificationFailed_subject: Message{
		en: "Verification failed",
		fr: "Échec de la vérification",
	},
	VerificationFailed_body: Message{
		en: "Your email cannot be added to the subscripted list, contact list owner for more infos.",
		fr: "Votre email ne peut pas être inscrit à la liste, veuillez contacter le propriétaire de la liste pour plus d'infos.",
	},
}
