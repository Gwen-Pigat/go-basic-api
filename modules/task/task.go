package task

import (
	"net/http"
	"oristack/initializers"
	"time"

	"github.com/go-chi/chi/v5"
)

type Task struct {
	ID      int     `json:"id"`
	Title   string  `json:"title"`
	DateAdd *string `json:"dateAdd"`
	DateTo  *string `json:"dateTo"`
	Content *string `json:"content"`
	IsDone  bool    `json:"isDone"`
	RefUser int     `json:"refUser"`
}

var taskSetup = map[string]string{
	"payload": "id,date_add,date_to,title,content,is_done,ref_user",
	"table":   "task",
}

func CreateTask(wrapper *initializers.Wrapper) {
	if err := wrapper.WrapData("title"); err != nil {
		wrapper.Error(err.Error())
		return
	}
	task := Task{
		Title:   wrapper.Data["title"].(string),
		IsDone:  false,
		RefUser: wrapper.ReturnUser(),
		DateAdd: initializers.StringPtr(time.Now().UTC().Truncate(time.Second).Format(initializers.DateFormat)),
	}
	smtp, err := initializers.DB.Prepare("INSERT INTO " + taskSetup["table"] + "(title,date_add,is_done,ref_user) VALUES(?,?,?,?)")
	if err != nil {
		wrapper.Error(err.Error(), 400)
		return
	}
	defer smtp.Close()
	_, err = smtp.Exec(task.Title, task.DateAdd, task.IsDone, task.RefUser)
	if err != nil {
		wrapper.Error(err.Error(), 400)
		return
	}
	GetTasks(wrapper)
}

func GetTasks(wrapper *initializers.Wrapper) {
	rows, err := initializers.DB.Query("SELECT "+taskSetup["payload"]+" FROM "+taskSetup["table"]+" WHERE ref_user=? ORDER BY date_add DESC", wrapper.ReturnUser())
	if err != nil {
		wrapper.Error(err.Error(), http.StatusInternalServerError)
		return
	}
	data := []Task{}
	for rows.Next() {
		var task Task
		if err := rows.Scan(&task.ID, &task.DateAdd, &task.DateTo, &task.Title, &task.Content, &task.IsDone, &task.RefUser); err != nil {
			wrapper.Error(err.Error(), http.StatusBadGateway)
			return
		}
		if task.DateTo != nil {
			*task.DateTo, err = initializers.WrapFormat(task.DateTo)
			if err != nil {
				wrapper.Error("Error parsing dateTp : " + err.Error())
				return
			}
		}
		*task.DateAdd, err = initializers.WrapFormat(task.DateAdd)
		if err != nil {
			wrapper.Error("Error parsing dateAdd : " + err.Error())
			return
		}
		data = append(data, task)
	}
	wrapper.Render(map[string]any{
		"data": data,
	}, 200)
}

func GetTask(wrapper *initializers.Wrapper) {
	rows, err := initializers.DB.Query("SELECT "+taskSetup["payload"]+" FROM "+taskSetup["table"]+" WHERE id=? ORDER BY date_add DESC", chi.URLParam(wrapper.Request, "id"))
	if err != nil {
		wrapper.Error(err.Error(), http.StatusInternalServerError)
		return
	}
	task := Task{}
	for rows.Next() {
		if err := rows.Scan(&task.ID, &task.DateAdd, &task.DateTo, &task.Title, &task.Content, &task.IsDone); err != nil {
			wrapper.Error(err.Error(), http.StatusBadGateway)
			return
		}
		if task.DateTo == nil {
			*task.DateTo = ""
		}
	}
	wrapper.Render(map[string]any{
		"task": task,
	}, 200)
}

func PatchTask(wrapper *initializers.Wrapper) {
	rows, err := initializers.DB.Query("SELECT "+taskSetup["payload"]+" FROM "+taskSetup["table"]+" WHERE id=? ORDER BY date_add DESC", chi.URLParam(wrapper.Request, "id"))
	if err != nil {
		wrapper.Error(err.Error(), http.StatusInternalServerError)
		return
	}
	task := Task{}
	for rows.Next() {
		if err := rows.Scan(&task.ID, &task.DateAdd, &task.DateTo, &task.Title, &task.Content, &task.IsDone, &task.RefUser); err != nil {
			wrapper.Error(err.Error(), http.StatusBadGateway)
			return
		}
	}
	task.DateTo = nil
	if !task.IsDone {
		task.DateTo = initializers.StringPtr(time.Now().UTC().Truncate(time.Second).Format(initializers.DateFormat))
	}
	task.IsDone = !task.IsDone
	rows, err = initializers.DB.Query(
		"UPDATE "+taskSetup["table"]+" SET is_done = ?,date_to=? WHERE id=? AND ref_user=?",
		task.IsDone, task.DateTo, chi.URLParam(wrapper.Request, "id"), wrapper.ReturnUser(),
	)
	if err != nil {
		wrapper.Error(err.Error(), http.StatusBadRequest)
		return
	}
	defer rows.Close()

	if task.DateTo != nil {
		*task.DateTo, err = initializers.WrapFormat(task.DateTo)
		if err != nil {
			wrapper.Error("Error parsing dateTo inside PATCH : " + err.Error())
			return
		}
	}

	wrapper.Render(map[string]any{
		"message": "Update successfull",
		"result":  task,
	})
}

func DeleteTask(wrapper *initializers.Wrapper) {
	rows, err := initializers.DB.Exec(
		"DELETE FROM "+taskSetup["table"]+" WHERE id=? AND ref_user=?",
		chi.URLParam(wrapper.Request, "id"), wrapper.ReturnUser(),
	)
	if err != nil {
		wrapper.Error(err.Error(), http.StatusBadRequest)
		return
	}
	if _, err := rows.RowsAffected(); err != nil {
		wrapper.Error(err.Error(), http.StatusBadRequest)
		return
	}
	wrapper.Render(map[string]any{
		"message": "Delete successfull",
	})
}
