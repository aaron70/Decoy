package random

import (
	_ "embed"
	"maps"
	"math/rand/v2"
	"slices"
	"strings"

	"github.com/aaron70/goaty/errors"
)

//go:embed corpus.txt
var DefaultRandomTextCorpus string

type RandomText struct {
	Rand           *rand.Rand
	n              int
	ngrams         map[string][]string
	keys           []string
}

func NewRandomText(rand *rand.Rand, n int) *RandomText {
	return &RandomText{
		n:      n,
		Rand:   rand,
		ngrams: make(map[string][]string),
	}
}

func (r *RandomText) SetNgrams(n int, ngrams map[string][]string) {
	r.n = n
	r.ngrams = ngrams
}

func (r *RandomText) NgramsFromString(text string) {
	words := strings.Fields(text)
	r.NgramsFromWords(words)
}

func (r *RandomText) NgramsFromWords(words []string) {
	if len(words) <= r.n {
		return
	}
	if len(r.ngrams) != 0 {
		r.ngrams = make(map[string][]string)
	}
	for i := 0; i < len(words)-r.n; i++ {
		key := strings.Join(words[i:i+r.n], " ")
		r.ngrams[key] = append(r.ngrams[key], words[i+r.n])
	}
	r.keys = slices.Collect(maps.Keys(r.ngrams))
}

func (r *RandomText) RandomText(maxWords int) (string, error) {
	if len(r.ngrams) == 0 {
		return "", errors.New("Ngrams not set")
	}

	words := strings.Fields(RandomChoice(r.Rand, r.keys...))
	text := ""
	for range maxWords {
		key := strings.Join(words[len(words)-r.n:], " ")
		successors, ok := r.ngrams[key]
		if !ok {
			break
		}
		words = append(words, RandomChoice(r.Rand, successors...))
		if text != "" {
			text += " "
		}
		text += words[len(words)-1]
	}

	return text, nil
}

func (r *RandomText) RandomName() string {
	return RandomChoice(r.Rand, firstNames...)
}

func (r *RandomText) RandomLastName() string {
	return RandomChoice(r.Rand, lastNames...)
}


var firstNames = []string{
	// English
	"James", "John", "Robert", "Michael", "William", "David", "Richard", "Joseph",
	"Thomas", "Charles", "Christopher", "Daniel", "Matthew", "Anthony", "Mark",
	"Donald", "Steven", "Paul", "Andrew", "Joshua", "Kenneth", "Kevin", "Brian",
	"George", "Timothy", "Ronald", "Edward", "Jason", "Jeffrey", "Ryan",
	"Jacob", "Gary", "Nicholas", "Eric", "Jonathan", "Stephen", "Larry",
	"Justin", "Scott", "Brandon", "Benjamin", "Samuel", "Raymond", "Gregory",
	"Frank", "Alexander", "Patrick", "Jack", "Dennis", "Jerry",
	// Female English
	"Mary", "Patricia", "Jennifer", "Linda", "Barbara", "Elizabeth", "Susan",
	"Jessica", "Sarah", "Karen", "Lisa", "Nancy", "Betty", "Margaret", "Sandra",
	"Ashley", "Dorothy", "Kimberly", "Emily", "Donna", "Michelle", "Carol",
	"Amanda", "Melissa", "Deborah", "Stephanie", "Rebecca", "Sharon", "Laura",
	"Cynthia", "Kathleen", "Amy", "Angela", "Shirley", "Anna", "Brenda",
	"Pamela", "Emma", "Nicole", "Helen", "Samantha", "Katherine", "Christine",
	"Debra", "Rachel", "Carolyn", "Janet", "Catherine", "Maria", "Heather",
	// Spanish/Latin
	"Alejandro", "Carlos", "Diego", "Eduardo", "Fernando", "Gabriel", "Héctor",
	"Ignacio", "Javier", "Luis", "Manuel", "Nicolás", "Oscar", "Pedro", "Rafael",
	"Sebastián", "Tomás", "Valentín", "Andrés", "Emilio", "Felipe", "Gerardo",
	"Sofía", "Valentina", "Isabella", "Camila", "Lucía", "Mariana", "Daniela",
	"Fernanda", "Gabriela", "Andrea", "Natalia", "Paula", "Rebeca", "Adriana",
	// French
	"Antoine", "Baptiste", "Clément", "Damien", "Étienne", "François", "Guillaume",
	"Hugo", "Julien", "Laurent", "Maxime", "Nicolas", "Olivier", "Pierre", "Quentin",
	"Raphaël", "Simon", "Théo", "Victor", "Xavier",
	"Amélie", "Camille", "Charlotte", "Claire", "Elise", "Emma", "Inès",
	"Julie", "Léa", "Lucie", "Manon", "Marie", "Mathilde", "Pauline", "Sophie",
	// German
	"Hans", "Klaus", "Werner", "Dieter", "Günter", "Helmut", "Rolf", "Gerhard",
	"Friedrich", "Heinz", "Stefan", "Matthias", "Andreas", "Markus", "Tobias",
	"Sabine", "Monika", "Ursula", "Ingrid", "Gisela", "Helga", "Petra", "Renate",
	"Brigitte", "Hildegard", "Anja", "Birgit", "Claudia", "Franziska", "Jana",
	// Italian
	"Marco", "Luca", "Giovanni", "Andrea", "Matteo", "Lorenzo", "Davide",
	"Alessandro", "Francesco", "Antonio", "Stefano", "Angelo", "Sergio", "Enrico",
	"Giulia", "Martina", "Sara", "Laura", "Francesca", "Silvia", "Valentina",
	"Chiara", "Alessia", "Roberta", "Federica", "Paola", "Elena", "Elisa",
	// Japanese (romanized)
	"Hiroshi", "Kenji", "Takashi", "Yoshio", "Akira", "Satoshi", "Daisuke",
	"Yuki", "Haruto", "Ren", "Sota", "Hayato", "Yuto", "Koki", "Ryota",
	"Yui", "Hana", "Sakura", "Aoi", "Rin", "Misaki", "Haruka", "Yuna",
	"Mei", "Nana", "Mio", "Saki", "Ayane", "Koharu",
	// Chinese (romanized)
	"Wei", "Fang", "Ming", "Lei", "Jun", "Jing", "Hui", "Xin", "Yi", "Hao",
	"Tao", "Ling", "Bin", "Chao", "Gang", "Peng", "Qiang", "Rui", "Tian",
	// Arabic
	"Ahmed", "Mohamed", "Ali", "Omar", "Hassan", "Ibrahim", "Khalid", "Tariq",
	"Youssef", "Samir", "Karim", "Nasser", "Bilal", "Hamid", "Jamal",
	"Fatima", "Aisha", "Layla", "Nour", "Rania", "Dina", "Hana", "Lina",
	// Indian
	"Rahul", "Amit", "Raj", "Sanjay", "Vikram", "Arjun", "Kiran", "Priya",
	"Ravi", "Suresh", "Anil", "Deepak", "Nikhil", "Rohit", "Vijay",
	"Priya", "Pooja", "Neha", "Anjali", "Kavita", "Sunita", "Meera", "Divya",
	// Russian
	"Ivan", "Dmitri", "Sergei", "Alexei", "Nikolai", "Pavel", "Mikhail",
	"Vladimir", "Andrei", "Boris", "Konstantin", "Yuri", "Oleg", "Viktor",
	"Olga", "Natasha", "Tatiana", "Elena", "Irina", "Ekaterina", "Svetlana",
	"Anna", "Marina", "Oksana", "Yulia", "Vera", "Galina",
	// Scandinavian
	"Lars", "Erik", "Bjorn", "Sven", "Henrik", "Magnus", "Olaf", "Nils",
	"Astrid", "Ingrid", "Freya", "Sigrid", "Helga", "Birgit", "Solveig",
	// Portuguese/Brazilian
	"João", "Miguel", "Rodrigo", "Gustavo", "Bruno", "Lucas", "Thiago",
	"Leonardo", "Rafael", "Mateus", "Igor", "Vinicius", "Caio", "Leandro",
	"Ana", "Beatriz", "Letícia", "Juliana", "Isabela", "Larissa", "Vanessa",
	"Fernanda", "Aline", "Camila", "Bruna", "Tatiane", "Renata",
	// Korean (romanized)
	"Minho", "Jungkook", "Taehyung", "Jimin", "Seokjin", "Namjoon", "Yoongi",
	"Junho", "Sehun", "Chanyeol", "Baekhyun", "Suho", "Hyunjin",
	"Jiyeon", "Sooyeon", "Hyuna", "Jisoo", "Jennie", "Lisa", "Rosé",
	// African names
	"Kwame", "Kofi", "Yaw", "Kojo", "Ama", "Abena", "Akua", "Adwoa",
	"Emeka", "Chukwu", "Obiora", "Nnamdi", "Eze", "Adaeze", "Ngozi",
	"Oluwaseun", "Babatunde", "Adebayo", "Taiwo", "Kehinde",
	"Sipho", "Thabo", "Bongani", "Lerato", "Nomsa", "Zanele",
}
 
var lastNames = []string{
	// English
	"Smith", "Johnson", "Williams", "Brown", "Jones", "Garcia", "Miller",
	"Davis", "Rodriguez", "Martinez", "Hernandez", "Lopez", "Gonzalez",
	"Wilson", "Anderson", "Thomas", "Taylor", "Moore", "Jackson", "Martin",
	"Lee", "Perez", "Thompson", "White", "Harris", "Sanchez", "Clark",
	"Ramirez", "Lewis", "Robinson", "Walker", "Young", "Allen", "King",
	"Wright", "Scott", "Torres", "Nguyen", "Hill", "Flores", "Green",
	"Adams", "Nelson", "Baker", "Hall", "Rivera", "Campbell", "Mitchell",
	"Carter", "Roberts", "Gomez", "Phillips", "Evans", "Turner", "Diaz",
	"Parker", "Cruz", "Edwards", "Collins", "Reyes", "Stewart", "Morris",
	"Morales", "Murphy", "Cook", "Rogers", "Gutierrez", "Ortiz", "Morgan",
	"Cooper", "Peterson", "Bailey", "Reed", "Kelly", "Howard", "Ramos",
	"Kim", "Cox", "Ward", "Richardson", "Watson", "Brooks", "Chavez",
	"Wood", "James", "Bennett", "Gray", "Mendoza", "Ruiz", "Hughes",
	"Price", "Alvarez", "Castillo", "Sanders", "Patel", "Myers", "Long",
	"Ross", "Foster", "Jimenez", "Powell", "Jenkins", "Perry", "Russell",
	"Sullivan", "Bell", "Coleman", "Butler", "Henderson", "Barnes", "Gonzales",
	"Fisher", "Vasquez", "Simmons", "Romero", "Jordan", "Patterson", "Alexander",
	"Hamilton", "Graham", "Reynolds", "Griffin", "Wallace", "Moreno", "West",
	"Cole", "Hayes", "Bryant", "Herrera", "Gibson", "Ellis", "Tran",
	// European
	"Müller", "Schmidt", "Schneider", "Fischer", "Weber", "Meyer", "Wagner",
	"Becker", "Schulz", "Hoffmann", "Schäfer", "Koch", "Bauer", "Richter",
	"Klein", "Wolf", "Schröder", "Neumann", "Schwarz", "Zimmermann",
	"Martin", "Bernard", "Dubois", "Thomas", "Robert", "Richard", "Petit",
	"Durand", "Leroy", "Moreau", "Simon", "Laurent", "Lefebvre", "Michel",
	"Garcia", "David", "Bertrand", "Roux", "Vincent", "Fournier",
	"Rossi", "Ferrari", "Esposito", "Bianchi", "Romano", "Colombo",
	"Ricci", "Marino", "Greco", "Bruno", "Gallo", "Conti", "De Luca",
	"Costa", "Giordano", "Mancini", "Rizzo", "Lombardi", "Moretti",
	"García", "Rodríguez", "Martínez", "López", "Sánchez", "Pérez",
	"González", "Fernández", "Romero", "Álvarez", "Torres", "Ramírez",
	"Flores", "Díaz", "Hernández", "Jiménez", "Morales", "Muñoz",
	"Silva", "Santos", "Ferreira", "Oliveira", "Souza", "Carvalho",
	"Almeida", "Nascimento", "Lima", "Araújo", "Ribeiro", "Rodrigues",
	// Slavic
	"Novak", "Kovač", "Horvat", "Babić", "Marić", "Nikolić", "Petrović",
	"Jovanović", "Popović", "Đorđević", "Stanković", "Marković", "Ilić",
	"Pavlović", "Milošević", "Vuković", "Kovačević", "Bogdanović",
	"Ivanov", "Petrov", "Sidorov", "Smirnov", "Volkov", "Sokolov",
	"Mikhailov", "Novikov", "Fedorov", "Morozov", "Volkov", "Alekseev",
	// Asian
	"Wang", "Li", "Zhang", "Liu", "Chen", "Yang", "Huang", "Zhao",
	"Wu", "Zhou", "Xu", "Sun", "Ma", "Zhu", "Hu", "Guo", "He", "Lin",
	"Tanaka", "Suzuki", "Sato", "Watanabe", "Ito", "Yamamoto", "Nakamura",
	"Kobayashi", "Kato", "Yoshida", "Yamada", "Sasaki", "Yamaguchi",
	"Kim", "Lee", "Park", "Choi", "Jung", "Kang", "Cho", "Yoon",
	"Nguyen", "Tran", "Le", "Pham", "Hoang", "Phan", "Vu", "Dang",
	// Indian
	"Patel", "Sharma", "Singh", "Kumar", "Verma", "Gupta", "Joshi",
	"Mishra", "Rao", "Nair", "Pillai", "Iyer", "Reddy", "Naidu",
	"Mehta", "Shah", "Chopra", "Malhotra", "Kapoor", "Bose", "Ghosh",
	// Middle Eastern
	"Al-Farsi", "Al-Hassan", "Al-Rashid", "Al-Mahmoud", "Al-Khalil",
	"Al-Sayed", "Al-Amin", "Al-Mansouri", "Al-Zahra", "Al-Qasim",
	"Abdullah", "Ibrahim", "Khalil", "Mustafa", "Hassan", "Hussein",
	// African
	"Okafor", "Okonkwo", "Adeyemi", "Abiodun", "Adeleke", "Adesanya",
	"Eze", "Nwosu", "Chukwu", "Obi", "Nwachukwu", "Uzor",
	"Nkosi", "Dlamini", "Khumalo", "Mthembu", "Ndlovu", "Zulu",
	"Boateng", "Mensah", "Asante", "Owusu", "Agyei", "Amponsah",
	// Scandinavian
	"Andersen", "Christensen", "Jensen", "Larsen", "Nielsen", "Pedersen",
	"Eriksson", "Karlsson", "Nilsson", "Persson", "Johansson", "Lindgren",
	"Berg", "Holm", "Strand", "Dahl", "Lund", "Vik", "Bakke",
}
