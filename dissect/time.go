package dissect

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
)

func readTime(r *Reader) error {
	time, err := r.Uint32()
	if err != nil {
		return err
	}
	r.time = float64(time)
	r.timeRaw = fmt.Sprintf("%d:%02d", time/60, time%60)
	return nil
}

func readY7Time(r *Reader) error {
	time, err := r.String()
	parts := strings.Split(time, ":")
	if len(parts) == 1 {
		seconds, err := strconv.ParseFloat(parts[0], 64)
		if err != nil {
			return err
		}
		r.time = seconds
		r.timeRaw = parts[0]
		return nil
	}
	minutes, err := strconv.Atoi(parts[0])
	if err != nil {
		return err
	}
	seconds, err := strconv.Atoi(parts[1])
	if err != nil {
		return err
	}
	r.time = float64((minutes * 60) + seconds)
	r.timeRaw = time
	return nil
}

func (r *Reader) roundEnd() {
	log.Debug().Msg("round_end")

	planter := -1
	deaths := make(map[int]int)
	sizes := make(map[int]int)
	roles := make(map[int]TeamRole)

	for _, p := range r.Header.Players {
		sizes[p.TeamIndex] += 1
		roles[p.TeamIndex] = r.Header.Teams[p.TeamIndex].Role
	}

	if r.Header.CodeVersion >= Y9S4 {
		team0Won := r.Header.Teams[0].StartingScore < r.Header.Teams[0].Score
		r.Header.Teams[0].Won = team0Won
		r.Header.Teams[1].Won = !team0Won
	}

	// ⚠ Имя из киллфида может НЕ найтись среди игроков: пакет игрока мог не
	// разобраться (сборка Y11S2_Alpha03), а индекс -1 роняет весь разбор
	// паникой. Раньше это не всплывало только потому, что проход обрывался
	// раньше и до roundEnd дело не доходило.
	team := func(name string) int {
		i := r.PlayerIndexByUsername(name)
		if i < 0 {
			log.Debug().Str("username", name).Msg("feedback about unknown player — skipped")
			return -1
		}
		return r.Header.Players[i].TeamIndex
	}
	for _, u := range r.MatchFeedback {
		switch u.Type {
		case Kill:
			i := team(u.Target)
			if i < 0 {
				break
			}
			deaths[i] = deaths[i] + 1
			// fix killer username
			if len(u.usernameFromScoreboard) > 0 {
				u.Username = u.usernameFromScoreboard
			}
			break
		case Death:
			i := team(u.Username)
			if i < 0 {
				break
			}
			deaths[i] = deaths[i] + 1
			break
		case DefuserPlantComplete:
			planter = r.PlayerIndexByUsername(u.Username)
			break
		case DefuserDisableComplete:
			i := team(u.Username)
			if i < 0 {
				break
			}
			r.Header.Teams[i].Won = true
			r.Header.Teams[i].WinCondition = DisabledDefuser
			return
		}
	}

	if planter > -1 {
		r.Header.Teams[r.Header.Players[planter].TeamIndex].Won = true
		r.Header.Teams[r.Header.Players[planter].TeamIndex].WinCondition = DefusedBomb
		return
	}

	// skip for now until we have a more reliable way of determining the win condition
	// Y9S4 at least tells us who won now in the header with StartingScore
	if r.Header.CodeVersion >= Y9S4 {
		return
	}

	if deaths[0] == sizes[0] {
		if planter > -1 && roles[0] == Attack { // ignore attackers killed post-plant
			return
		}
		r.Header.Teams[1].Won = true
		r.Header.Teams[1].WinCondition = KilledOpponents
		return
	}
	if deaths[1] == sizes[1] {
		if planter > -1 && roles[1] == Attack { // ignore attackers killed post-plant
			return
		}
		r.Header.Teams[0].Won = true
		r.Header.Teams[0].WinCondition = KilledOpponents
		return
	}

	i := 0
	if roles[1] == Defense {
		i = 1
	}

	r.Header.Teams[i].Won = true
	r.Header.Teams[i].WinCondition = Time
}
