package handler

import (
	"net/http"

	"github.com/ABDELRAHMAN-ELRAYES/Vai/pkg/apierror"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/pkg/httputil"
)

func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"status":  "ok",
		"env":     h.App.Config.Env,
		"version": "0.0.1",
	}

	if  err :=httputil.JSONResponse(w, http.StatusOK, data);err != nil{
		apierror.InternalServerError(h.App.Logger,w,r,err)
		return
	}

}
