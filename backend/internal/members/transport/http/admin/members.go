package admin

import (
	"net/http"

	apphttp "expensebot/internal/app/http"
)

func (h *Handlers) getMembers(w http.ResponseWriter, r *http.Request) {
	list, err := h.membersUC.GetMembers(r.Context())
	if err != nil {
		apphttp.Fail(w, h.log, err)
		return
	}

	out := make([]memberOut, 0, len(list))
	for _, m := range list {
		out = append(out, mapMemberToOut(m))
	}

	apphttp.WriteJSON(w, http.StatusOK, out)
}

func (h *Handlers) updateMember(w http.ResponseWriter, r *http.Request) {
	id, ok := apphttp.PathID(w, r)
	if !ok {
		return
	}

	var in updateMemberIn
	if !apphttp.Decode(w, r, &in) {
		return
	}

	m, err := h.membersUC.UpdateMember(r.Context(), id, mapUpdateIn(in))
	if err != nil {
		apphttp.Fail(w, h.log, err)
		return
	}

	apphttp.WriteJSON(w, http.StatusOK, mapMemberToOut(*m))
}
